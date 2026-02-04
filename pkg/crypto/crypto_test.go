package crypto

import (
	"context"
	"os"
	"testing"
)

// 测试AES-CBC模式
func TestAESCipher(t *testing.T) {
	ctx := context.Background()
	key := []byte("your-secret-key1") // 16 bytes for AES-128

	// 创建 AES 加密器
	aesCipher, err := NewAESCipher(key)
	if err != nil {
		t.Fatalf("创建AES加密器错误: %+v", err)
	}

	// 测试加密解密
	plaintext := "Hello, World!"
	ciphertext, err := aesCipher.Encrypt(ctx, plaintext)
	if err != nil {
		t.Fatalf("加密错误: %+v", err)
	}

	decrypted, err := aesCipher.Decrypt(ctx, ciphertext)
	if err != nil {
		t.Fatalf("解密错误: %+v", err)
	}

	if decrypted != plaintext {
		t.Errorf("解密文本与原始文本不匹配: 得到 %s, 期望 %s", decrypted, plaintext)
	}
}

// 测试DES模式
func TestDESCipher(t *testing.T) {
	ctx := context.Background()
	key := []byte("your-key") // 8 bytes for DES

	// 创建 DES 加密器
	desCipher, err := NewDESCipher(key)
	if err != nil {
		t.Fatalf("创建DES加密器错误: %+v", err)
	}

	// 测试加密解密
	plaintext := "Hello, World!"
	ciphertext, err := desCipher.Encrypt(ctx, plaintext)
	if err != nil {
		t.Fatalf("加密错误: %+v", err)
	}

	decrypted, err := desCipher.Decrypt(ctx, ciphertext)
	if err != nil {
		t.Fatalf("解密错误: %+v", err)
	}

	if decrypted != plaintext {
		t.Errorf("解密文本与原始文本不匹配: 得到 %s, 期望 %s", decrypted, plaintext)
	}
}

// 测试SHA256哈希
func TestSHA256Hasher(t *testing.T) {
	ctx := context.Background()
	hasher := NewSHA256Hasher()

	// 测试哈希和验证
	data := "Hello, World!"
	hash, err := hasher.Hash(ctx, data)
	if err != nil {
		t.Fatalf("哈希错误: %+v", err)
	}

	valid, err := hasher.Verify(ctx, data, hash)
	if err != nil {
		t.Fatalf("验证错误: %+v", err)
	}

	if !valid {
		t.Error("哈希验证失败")
	}

	// 测试验证失败的情况
	invalidValid, err := hasher.Verify(ctx, "invalid data", hash)
	if err != nil {
		t.Fatalf("验证无效数据错误: %+v", err)
	}

	if invalidValid {
		t.Error("无效数据的哈希验证应该失败")
	}
}

// 测试SHA512哈希
func TestSHA512Hasher(t *testing.T) {
	ctx := context.Background()
	hasher := NewSHA512Hasher()

	// 测试哈希和验证
	data := "Hello, World!"
	hash, err := hasher.Hash(ctx, data)
	if err != nil {
		t.Fatalf("哈希错误: %+v", err)
	}

	valid, err := hasher.Verify(ctx, data, hash)
	if err != nil {
		t.Fatalf("验证错误: %+v", err)
	}

	if !valid {
		t.Error("哈希验证失败")
	}
}

// 测试Bcrypt哈希
func TestBcryptHasher(t *testing.T) {
	ctx := context.Background()
	hasher := NewBcryptHasher(0) // 使用默认成本

	// 测试哈希和验证
	password := "my-secret-password"
	hash, err := hasher.Hash(ctx, password)
	if err != nil {
		t.Fatalf("哈希错误: %+v", err)
	}

	valid, err := hasher.Verify(ctx, password, hash)
	if err != nil {
		t.Fatalf("验证错误: %+v", err)
	}

	if !valid {
		t.Error("哈希验证失败")
	}

	// 测试验证失败的情况
	invalidValid, err := hasher.Verify(ctx, "invalid password", hash)
	if err != nil {
		t.Fatalf("验证无效密码错误: %+v", err)
	}

	if invalidValid {
		t.Error("无效密码的哈希验证应该失败")
	}
}

// 测试Scrypt哈希
func TestScryptHasher(t *testing.T) {
	ctx := context.Background()
	hasher := NewScryptHasher()

	// 测试哈希和验证
	data := "Hello, World!"
	hash, err := hasher.Hash(ctx, data)
	if err != nil {
		t.Fatalf("哈希错误: %+v", err)
	}

	valid, err := hasher.Verify(ctx, data, hash)
	if err != nil {
		t.Fatalf("验证错误: %+v", err)
	}

	if !valid {
		t.Error("哈希验证失败")
	}

	// 测试验证失败的情况
	invalidValid, err := hasher.Verify(ctx, "invalid data", hash)
	if err != nil {
		t.Fatalf("验证无效数据错误: %+v", err)
	}

	if invalidValid {
		t.Error("无效数据的哈希验证应该失败")
	}
}

// 测试上下文取消
func TestContextCancellation(t *testing.T) {
	// 创建一个已取消的 context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	// 测试 AES 加密
	key := []byte("your-secret-key1")
	aesCipher, err := NewAESCipher(key)
	if err != nil {
		t.Fatalf("创建AES加密器错误: %+v", err)
	}

	_, err = aesCipher.Encrypt(ctx, "Hello, World!")
	if err == nil {
		t.Error("期望取消上下文时返回错误，但得到nil")
	}

	// 测试 SHA256 哈希
	hasher := NewSHA256Hasher()
	_, err = hasher.Hash(ctx, "Hello, World!")
	if err == nil {
		t.Error("期望取消上下文时返回错误，但得到nil")
	}
}

// 测试AES-GCM模式
func TestAESGCMCipher(t *testing.T) {
	ctx := context.Background()
	key := []byte("your-secret-key12345678901234567") // 32 bytes for AES-256

	// 创建 AES-GCM 加密器
	aesGCMCipher, err := NewAESGCMCipher(key)
	if err != nil {
		t.Fatalf("创建AES-GCM加密器错误: %+v", err)
	}

	// 测试加密解密
	plaintext := "Hello, World!"
	ciphertext, err := aesGCMCipher.Encrypt(ctx, plaintext)
	if err != nil {
		t.Fatalf("加密错误: %+v", err)
	}

	decrypted, err := aesGCMCipher.Decrypt(ctx, ciphertext)
	if err != nil {
		t.Fatalf("解密错误: %+v", err)
	}

	if decrypted != plaintext {
		t.Errorf("解密文本与原始文本不匹配: 得到 %s, 期望 %s", decrypted, plaintext)
	}
}

// 测试密钥管理功能
func TestKeyManager(t *testing.T) {
	// 创建密钥管理器
	km := NewKeyManager()

	// 测试从密码派生密钥
	password := "my-secret-password"
	key, err := km.DeriveKey(password, nil)
	if err != nil {
		t.Fatalf("派生密钥错误: %+v", err)
	}

	if len(key) != 32 {
		t.Errorf("派生密钥长度与预期不符: 得到 %d, 期望 %d", len(key), 32)
	}

	// 测试生成随机密钥
	randomKey, err := km.GenerateRandomKey()
	if err != nil {
		t.Fatalf("生成随机密钥错误: %+v", err)
	}

	if len(randomKey) != 32 {
		t.Errorf("生成的随机密钥长度与预期不符: 得到 %d, 期望 %d", len(randomKey), 32)
	}

	// 测试密钥格式转换
	keyStr := KeyToBase64(key)
	decodedKey, err := KeyFromBase64(keyStr)
	if err != nil {
		t.Fatalf("从base64解码密钥错误: %+v", err)
	}

	if string(decodedKey) != string(key) {
		t.Error("解码后的密钥与原始密钥不匹配")
	}

	// 测试密钥大小验证
	err = ValidateKeySize(key, "aes")
	if err != nil {
		t.Fatalf("验证AES密钥大小错误: %+v", err)
	}
}

// 测试HMAC功能
func TestHMACHasher(t *testing.T) {
	ctx := context.Background()
	key := []byte("my-secret-key")

	// 测试HMAC-SHA256
	hmacHasher := NewHMACHasher(key, HMACSHA256)

	// 测试哈希和验证
	data := "Hello, World!"
	hash, err := hmacHasher.Hash(ctx, data)
	if err != nil {
		t.Fatalf("哈希错误: %+v", err)
	}

	valid, err := hmacHasher.Verify(ctx, data, hash)
	if err != nil {
		t.Fatalf("验证错误: %+v", err)
	}

	if !valid {
		t.Error("HMAC验证失败")
	}

	// 测试验证失败的情况
	invalidValid, err := hmacHasher.Verify(ctx, "invalid data", hash)
	if err != nil {
		t.Fatalf("验证无效数据错误: %+v", err)
	}

	if invalidValid {
		t.Error("无效数据的HMAC验证应该失败")
	}

	// 测试HMAC-SHA512
	hmacHasher512 := NewHMACHasher(key, HMACSHA512)
	hash512, err := hmacHasher512.Hash(ctx, data)
	if err != nil {
		t.Fatalf("使用SHA512哈希错误: %+v", err)
	}

	if len(hash512) == len(hash) {
		t.Error("HMAC-SHA512哈希长度应该与HMAC-SHA256不同")
	}
}

// 测试文件加密/解密功能
func TestFileEncryptor(t *testing.T) {
	ctx := context.Background()
	key := []byte("your-secret-key1") // 16 bytes for AES-128

	// 创建 AES 加密器
	aesCipher, err := NewAESCipher(key)
	if err != nil {
		t.Fatalf("创建AES加密器错误: %+v", err)
	}

	// 创建文件加密器
	fileEncryptor := NewAESFileEncryptor(aesCipher)

	// 创建测试文件
	testContent := "Hello, File Encryption!"
	srcPath := "/tmp/test.txt"
	dstPath := "/tmp/test.encrypted"
	decryptedPath := "/tmp/test.decrypted"

	// 确保测试文件在测试结束时被清理
	defer func() {
		os.Remove(srcPath)
		os.Remove(dstPath)
		os.Remove(decryptedPath)
	}()

	// 写入测试内容
	err = os.WriteFile(srcPath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("写入测试文件错误: %+v", err)
	}

	// 测试加密文件
	err = fileEncryptor.EncryptFile(ctx, srcPath, dstPath)
	if err != nil {
		t.Fatalf("加密文件错误: %+v", err)
	}

	// 测试解密文件
	err = fileEncryptor.DecryptFile(ctx, dstPath, decryptedPath)
	if err != nil {
		t.Fatalf("解密文件错误: %+v", err)
	}

	// 读取解密后的内容
	decryptedContent, err := os.ReadFile(decryptedPath)
	if err != nil {
		t.Fatalf("读取解密文件错误: %+v", err)
	}

	if string(decryptedContent) != testContent {
		t.Errorf("解密文件内容与原始内容不匹配: 得到 %s, 期望 %s", string(decryptedContent), testContent)
	}
}

// 测试工具函数
func TestUtils(t *testing.T) {
	// 测试随机字节生成
	randomBytes, err := GenerateRandomBytes(16)
	if err != nil {
		t.Fatalf("生成随机字节错误: %+v", err)
	}

	if len(randomBytes) != 16 {
		t.Errorf("生成的随机字节长度与预期不符: 得到 %d, 期望 %d", len(randomBytes), 16)
	}

	// 测试随机字符串生成
	randomString, err := GenerateRandomString(16)
	if err != nil {
		t.Fatalf("生成随机字符串错误: %+v", err)
	}

	if len(randomString) == 0 {
		t.Error("生成的随机字符串为空")
	}

	// 测试随机整数生成
	randomInt, err := GenerateRandomInt(1, 100)
	if err != nil {
		t.Fatalf("生成随机整数错误: %+v", err)
	}

	if randomInt < 1 || randomInt > 100 {
		t.Errorf("生成的随机整数超出范围: 得到 %d", randomInt)
	}

	// 测试编码/解码
	testData := []byte("Hello, World!")
	base64Str := EncodeBase64(testData)
	decodedData, err := DecodeBase64(base64Str)
	if err != nil {
		t.Fatalf("解码base64错误: %+v", err)
	}

	if string(decodedData) != string(testData) {
		t.Error("解码后的base64数据与原始数据不匹配")
	}

	// 测试安全比较
	if !SafeEqual("test", "test") {
		t.Error("SafeEqual应该对相等的字符串返回true")
	}

	if SafeEqual("test", "test1") {
		t.Error("SafeEqual应该对不相等的字符串返回false")
	}
}

// 测试边界情况
func TestEdgeCases(t *testing.T) {
	ctx := context.Background()

	// 测试空字符串加密
	key := []byte("your-secret-key1") // 16 bytes for AES-128
	aesCipher, err := NewAESCipher(key)
	if err != nil {
		t.Fatalf("创建AES加密器错误: %+v", err)
	}

	emptyString := ""
	encrypted, err := aesCipher.Encrypt(ctx, emptyString)
	if err != nil {
		t.Fatalf("加密空字符串错误: %+v", err)
	}

	decrypted, err := aesCipher.Decrypt(ctx, encrypted)
	if err != nil {
		t.Fatalf("解密空字符串错误: %+v", err)
	}

	if decrypted != emptyString {
		t.Error("解密的空字符串与原始字符串不匹配")
	}

	// 测试上下文取消
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // 立即取消

	_, err = aesCipher.Encrypt(cancelCtx, "Hello")
	if err == nil {
		t.Error("期望取消上下文时返回错误，但得到nil")
	}
}

// 性能测试
func BenchmarkAESCipher(b *testing.B) {
	ctx := context.Background()
	key := []byte("your-secret-key1") // 16 bytes for AES-128
	aesCipher, err := NewAESCipher(key)
	if err != nil {
		b.Fatalf("创建AES加密器错误: %+v", err)
	}

	testData := "Hello, World!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := aesCipher.Encrypt(ctx, testData)
		if err != nil {
			b.Fatalf("加密错误: %+v", err)
		}
	}
}

func BenchmarkAESGCMCipher(b *testing.B) {
	ctx := context.Background()
	key := []byte("your-secret-key12345678901234567") // 32 bytes for AES-256
	aesGCMCipher, err := NewAESGCMCipher(key)
	if err != nil {
		b.Fatalf("创建AES-GCM加密器错误: %+v", err)
	}

	testData := "Hello, World!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := aesGCMCipher.Encrypt(ctx, testData)
		if err != nil {
			b.Fatalf("加密错误: %+v", err)
		}
	}
}

func BenchmarkSHA256Hasher(b *testing.B) {
	ctx := context.Background()
	hasher := NewSHA256Hasher()

	testData := "Hello, World!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hasher.Hash(ctx, testData)
		if err != nil {
			b.Fatalf("哈希错误: %+v", err)
		}
	}
}

func BenchmarkHMACHasher(b *testing.B) {
	ctx := context.Background()
	key := []byte("my-secret-key")
	hasher := NewHMACHasher(key, HMACSHA256)

	testData := "Hello, World!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hasher.Hash(ctx, testData)
		if err != nil {
			b.Fatalf("哈希错误: %+v", err)
		}
	}
}
