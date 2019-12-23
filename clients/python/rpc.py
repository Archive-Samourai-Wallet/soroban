import time, json, requests, hashlib

class RpcTimeout(Exception):
    pass

class RpcCall(Exception):
    pass

def get_session(with_tor):
    session = requests.session()
    if with_tor:
        # Tor uses the 9050 port as the default socks port
        session.proxies = {'http' : 'socks5h://%s' % with_tor,
                           'https': 'socks5h://%s' % with_tor}
    return session

def encode_directory(name):
    if not name or not isinstance(name, str):
        raise Exception("encode_directory invalid name")
    return hashlib.sha256(name.encode('utf-8')).hexdigest()

class Rpc:
    def __init__(self, url, with_tor):
        self.url = url
        self.session = get_session(with_tor)

    def call(self, method, args):
        try:
            headers = {
                'content-type': 'application/json',
                'User-Agent': 'HotJava/1.1.2 FCS'
            }
            payload = {
                "method": method,
                "params": [args],
                "jsonrpc": "2.0",
                "id": 1,
            }
            response = self.session.post(self.url, data=json.dumps(payload), headers=headers).json()
            return response['result']
        except:
            raise RpcCall("RPC error: %s" % method)


    def directory_list(self, name):
        resp = self.call('directory.List', {'Name': name, 'Entries': []})
        return resp.get('Entries', [])

    def directory_add(self, name, entry, mode='default'):
        resp = self.call('directory.Add', {'Name': name, 'Entry': entry, 'Mode': mode})
        return resp.get('Status', "") not in ["success"]

    def directory_remove(self, name, entry):
        resp = self.call('directory.Remove', {'Name': name, 'Entry': entry})
        return resp.get('Status', "") not in ["success"]

    def wait_and_remove(self, directory, count=25):
        values = []
        total = count
        count = 0
        while count < total:
            values = self.directory_list(directory)
            count += 1
            if len(values) > 0 or count >= total:
                break
            # wait for next list
            time.sleep(0.2)

        if count >= total:
            raise RpcTimeout("Wait on %s" % directory[:8])

        print("Wait try count (%d/%d)" % (count+1, total))

        # consider last entry
        value = values[-1]
        self.directory_remove(directory, value)
        return value
