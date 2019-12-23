from user import User, box_encrypt, box_decrypt, box_shared
from rpc import Rpc, encode_directory

def demo(url, directory_name, hash_encode, rpc):
    ## Alice
    print()
    print("=== Alice ===")
    # generate private key
    alice = User()
    print("Alice:",  "pubKey", alice.public_key())

    print("Alice:", "Add candidate in directory")
    rpc.directory_add(directory_name, alice.public_key(), "long")

    ## Bob
    print()
    print("=== Bob ===")
    bob = User()

    # generate private key
    print("Bob:", "pubKey", bob.public_key())

    # wait for candidate
    candidate_publicKey = rpc.wait_and_remove(directory_name)

    print("Bob:", "Send pubkey to alice private directory")
    private_directory = directory_name+'.'+alice.public_key()
    if hash_encode:
        private_directory = encode_directory(private_directory)

    rpc.directory_add(private_directory, bob.public_key(), "default")

    # generate nacl box
    bob_box = bob.box(candidate_publicKey)
    # get shared directory    
    shared_directory = encode_directory(box_shared(bob_box))
    print("Bob:", "Shared directory", encode_directory(box_shared(bob_box)))

    ## Alice
    print()
    print("=== Alice ===")

    # wait for candidate
    candidate_publicKey = rpc.wait_and_remove(private_directory)

    # remove offer
    print("Alice:", "Remove candidate from directory")
    rpc.directory_remove(directory_name, alice.public_key())

    # generate nacl box
    alice_box = alice.box(candidate_publicKey)

    # get shared directory    
    shared_directory = encode_directory(box_shared(alice_box))
    print("Alice:", "Shared directory", shared_directory)

    payload = box_encrypt(alice_box, 'Hello From Alice')
    if not payload:
        return

    rpc.directory_add(shared_directory, payload, "short")

    alice_next_directory = encode_directory(payload)
    print("Alice:", "Next directory", alice_next_directory)

    print()
    print("=== Bob ===")
    # consider last response
    payload = rpc.wait_and_remove(shared_directory)

    message = box_decrypt(bob_box, payload)
    if not message:
        return
    response_directory = encode_directory(payload)
    print("Bob:", "Alice's message: ", message)

    payload = box_encrypt(alice_box, 'Hi from Bob')
    if not payload:
        return

    rpc.directory_add(response_directory, payload, "short")

    print()
    print("=== Alice ===")
    # consider last response
    payload = rpc.wait_and_remove(alice_next_directory)

    message = box_decrypt(alice_box, payload)
    if not message:
        return
    print("Alice:", "Bob's Response: ", message)
