import nacl.utils
from nacl.public import PublicKey, PrivateKey, Box


class User:
    def __init__(self):
        self.privateKey = PrivateKey.generate()
        self.publicKey = self.privateKey.public_key

    def public_key(self):
        return bytes(self.privateKey.public_key).hex()

    def box(self, publicKey):
        return Box(self.privateKey, PublicKey(bytes.fromhex(publicKey)))

def box_shared(box):
    return box.shared_key().hex()

def box_encrypt(box, message):
    if not message or not isinstance(message, str):
        raise Exception("box_encrypt invalid message")
    nonce = nacl.utils.random(Box.NONCE_SIZE)
    encrypted = box.encrypt(message.encode('utf-8'), nonce)
    return encrypted.hex()


def box_decrypt(box, data):
    if not data or not isinstance(data, str):
        raise Exception("box_decrypt invalid data")
    encrypted = bytes.fromhex(data)
    nonce = encrypted[:24]
    encrypted = encrypted[24:]
    return box.decrypt(encrypted, nonce).decode('utf-8')
