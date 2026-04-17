import base64

from gmssl.sm4 import (
    CryptSM4, 
    SM4_ENCRYPT, 
)

SECRET_KEY: bytes = b"8ea82cdb8bc5ca1f7337ee4db48d675d"
SECRET_IV: bytes = b"c4f847566596ce07"


def sm4_cbc_encrypt(original_str: str) -> str:
    """
    国密SM4加密函数。
    使用CBC模式对原始字符串进行加密, 并返回经过URL安全的Base64编码的密文。

    :param original_str: 加密前的原文字符串。
    :return: 经过URL安全Base64编码的加密文本。
    :raises RuntimeError: 当加密过程中遇到任何异常时抛出。
    """
    try:
        encode_bytes = base64.b64encode(original_str.encode("utf-8"))
        crypt_sm4 = CryptSM4()
        crypt_sm4.set_key(SECRET_KEY, SM4_ENCRYPT)
        en_text = crypt_sm4.crypt_cbc(SECRET_IV, encode_bytes)
        return base64.urlsafe_b64encode(en_text).decode("utf-8")
    except Exception as e:
        raise RuntimeError(f"数据加密失败: {e}")


print(sm4_cbc_encrypt("2029-04-12T23:59:59"))
