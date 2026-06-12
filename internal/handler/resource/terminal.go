package resource

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	resomodel "gin-artweb/internal/model/resource"
	"gin-artweb/internal/service/resource"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type TerminalHandler struct {
	logger      *zap.Logger
	termService *resource.TerminalService
}

func NewTerminalHandler(logger *zap.Logger, termService *resource.TerminalService) *TerminalHandler {
	return &TerminalHandler{
		logger:      logger,
		termService: termService,
	}
}

func (h *TerminalHandler) HandleWebSocket(c *gin.Context) {
	ctx := c.Request.Context()
	log := ctxutil.NewLogger(h.logger, ctx)

	var req resomodel.HostSSHDTO
	if !common.ShouldBindQuery(c, log, &req, "查询主机列表:绑定查询主机列表参数失败") {
		return
	}

	cols := req.Columns
	if cols <= 0 {
		cols = 80
	}

	rows := req.Rows
	if rows <= 0 {
		rows = 24
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("WebSocket升级失败", zap.Error(err))
		return
	}
	defer conn.Close()

	session, err := h.termService.CreateSessionByHostID(c.Request.Context(), req.HostID, "xterm-256color", cols, rows)
	if err != nil {
		h.logger.Error("创建终端会话失败", zap.Error(err), zap.Uint32("host_id", req.HostID))
		_ = conn.WriteMessage(websocket.TextMessage, []byte("连接失败: "+err.Error()))
		return
	}
	defer session.Close()

	if err := session.StartShell(); err != nil {
		h.logger.Error("启动Shell失败", zap.Error(err))
		_ = conn.WriteMessage(websocket.TextMessage, []byte("启动Shell失败: "+err.Error()))
		return
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			data, err := session.Read()
			if err != nil {
				if err.Error() != "EOF" {
					h.logger.Error("读取终端输出失败", zap.Error(err))
				}
				return
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				h.logger.Error("WebSocket写入失败", zap.Error(err))
				return
			}
		}
	}()

	go func() {
		for {
			data, err := session.ReadErr()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				return
			}
		}
	}()

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if !strings.Contains(err.Error(), "close") {
					h.logger.Error("WebSocket读取失败", zap.Error(err))
				}
				return
			}

			if len(message) > 0 && message[0] == '\x00' {
				parts := strings.SplitN(string(message[1:]), ",", 2)
				if len(parts) == 2 {
					if newCols, err := strconv.Atoi(parts[0]); err == nil {
						if newRows, err := strconv.Atoi(parts[1]); err == nil {
							_ = session.Resize(newCols, newRows)
						}
					}
				}
				continue
			}

			_, err = session.Write(message)
			if err != nil {
				h.logger.Error("写入终端失败", zap.Error(err))
				return
			}
		}
	}()

	<-done
}
