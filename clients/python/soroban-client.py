import sys, time, json, datetime
import base64
import traceback

from user import User, box_encrypt, box_decrypt, box_shared
from rpc import Rpc, encode_directory, RpcTimeout, RpcCall
from cli import parse as parse_cli


def initiator(url, directory_name, hash_encode, rpc, num_iter):
    user = User()
    print("Registering public_key")
    rpc.directory_add(directory_name, user.public_key(), "long")
    
    private_directory = "%s.%s" % (directory_name, user.public_key())
    if hash_encode:
        private_directory = encode_directory(private_directory)
    if not private_directory:
        print("Invalid private_directory")
        return

    candidate_publicKey = rpc.wait_and_remove(private_directory, 100)
    if not candidate_publicKey:
        print("Invalid candidate_publicKey")
        return
    print("Contributor public_key found")

    box = user.box(candidate_publicKey)
    next_directory = encode_directory(box_shared(box))
    if not next_directory:
        return
    payload = ""

    print("Starting echange loop...")
    counter = 1
    while num_iter>0:
        # request
        print("Sending :", "Ping")
        payload = box_encrypt(box, "Ping %d %s" % (counter, datetime.datetime.now()))
        if not payload:
            print("Invalid Ping message")
            return
        rpc.directory_add(next_directory, payload, "short")
        next_directory = encode_directory(payload)
        if not next_directory:
            print("Invalid reponse next_directory")
            return

        # response
        payload = rpc.wait_and_remove(next_directory, 10)
        next_directory = encode_directory(payload)
        if not next_directory:
            print("Invalid reponse next_directory")
            return

        message = box_decrypt(box, payload)
        if not message:
            print("Invalid reponse message")
            return
        print("Recieved:", message)
        counter += 1
        num_iter -= 1


def contributor(url, directory_name, hash_encode, rpc, num_iter):
    user = User()

    initiator_publicKey = rpc.wait_and_remove(directory_name, 10)
    if not initiator_publicKey:
        print("Invalid initiator_publicKey")
        return
    print("Initiator public_key found")

    private_directory = "%s.%s" % (directory_name, initiator_publicKey)
    if hash_encode:
        private_directory = encode_directory(private_directory)

    if not private_directory:
        print("Invalid private_directory")
        return

    print("Sending public_key")
    rpc.directory_add(private_directory, user.public_key(), "default")

    box = user.box(initiator_publicKey)
    next_directory = encode_directory(box_shared(box))
    if not next_directory:
        print("Invalid next_directory (start)")
        return
    payload = ""

    print("Starting echange loop...")
    counter = 1
    while num_iter>0:
        # query
        payload = rpc.wait_and_remove(next_directory, 10)
        if not payload:
            print("Invalid payload (query)")
            return
        next_directory = encode_directory(payload)
        if not next_directory:
            print("Invalid next_directory (query)", next_directory)
            return

        message = box_decrypt(box, payload)
        if not message:
            print("Invalid query")
            return
        print("Recieved:", message)

        # response
        print("Replying:", "Pong")
        payload = box_encrypt(box, "Pong %d %s" % (counter, datetime.datetime.now()))
        if not payload:
            print("Invalid payload (reply)")
            return
        rpc.directory_add(next_directory, payload, "short")

        next_directory = encode_directory(payload)
        if not next_directory:
            print("Invalid next_directory (response)")
            return

        counter += 1
        num_iter -= 1


def main(argv):
    url, with_tor, directory_name, hash_encode, role, num_iter = parse_cli(argv)
    if hash_encode:
        directory_name = encode_directory(directory_name)
    if not directory_name:
        return

    rpc = Rpc(url, with_tor)

    while True:
        try:
            if role=="initiator":
                initiator(url, directory_name, hash_encode, rpc, num_iter)
            elif role=="contributor":
                contributor(url, directory_name, hash_encode, rpc, num_iter)
            else:
                raise Exception("Invalid role")
            print("Done.")

        except RpcTimeout as e:
            traceback.print_exc()
            print("RpcTimeout:", e)
        except RpcCall as e:
            traceback.print_exc()
            print("RpcCall:", e)
        except Exception as e:
            traceback.print_exc()
            print("Error:", e)
        except:
            traceback.print_exc()
            print("Unknown error")


if __name__ == "__main__":
   main(sys.argv[1:])
