import sys, getopt


def parse(argv):
    url = 'http://localhost:4242'
    with_tor = "localhost:9150"
    directory_name = ''
    hash_encode = False
    role = "demo"
    num_iter = 3

    try:
        opts, args = getopt.getopt(argv,"heu:t:d:r:n:",["url=", "with_tor=", "directory_name=", "role=", "num_iter"])
    except getopt.GetoptError as e:
        print('sordoban -d <directory_name>', e)
        sys.exit(2)

    for opt, arg in opts:
        if opt == '-h':
            print('sordoban -d <directory_name>')
            sys.exit()
        if opt == '-e':
            hash_encode = True
        elif opt in ["-u", "--url"]:
            url = arg
        elif opt in ["-t", "--with_tor"]:
            with_tor = arg
        elif opt in ["-d", "--directory_name"]:
            directory_name = arg
        elif opt in ["-r", "--role"]:
            role = arg
        elif opt in ["-n", "--num_iter"]:
            num_iter = int(arg)

    return (url, with_tor, directory_name, hash_encode, role, num_iter)
