#!/bin/sh
set -euo pipefail 
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

usage () {
  echo "Soroban startup script
  $(basename "$0") command

  where:
      server commands [server_build server_start server_stop server_status server_logs]
      client commands [clients_build clients_start clients_stop clients_python_logs clients_java_logs]"
}

soroban () {
  COMMAND=${1-default}  

  case "$COMMAND" in
    # server
    "server_build")
        docker-compose -f ${DIR}/docker-compose.yml build;;

    "server_start")
        docker-compose -f ${DIR}/docker-compose.yml up -d;;

    "server_stop")
        docker-compose -f ${DIR}/docker-compose.yml down;;

    "server_status")
        watch docker-compose -f ${DIR}/docker-compose.yml ps;;

    "server_logs")
        docker-compose -f ${DIR}/docker-compose.yml logs -f;;

    "clients_build")
        docker-compose -f ${DIR}/clients/python/docker-compose.yml build
        docker-compose -f ${DIR}/clients/java/soroban-client/docker-compose.yml build
        ;;

    "clients_start")
        docker-compose -f ${DIR}/clients/python/docker-compose.yml up -d --scale initiator=4 --scale contributor=4
        docker-compose -f ${DIR}/clients/java/soroban-client/docker-compose.yml up -d --scale initiator=4 --scale contributor=4
        ;;

    "clients_stop")
        docker-compose -f ${DIR}/clients/python/docker-compose.yml down
        docker-compose -f ${DIR}/clients/java/soroban-client/docker-compose.yml down
        ;;

    "clients_python_logs")
        docker-compose -f ${DIR}/clients/python/docker-compose.yml logs -f --tail=100;;

    "clients_java_logs")
        docker-compose -f ${DIR}/clients/java/soroban-client/docker-compose.yml logs -f --tail=100;;

    *)
      usage
      exit 1;;
  esac

}

main () {
  if [  $# -le 0 ]; then 
    usage $@
    exit 1
  fi 

  soroban $@
  exit
}

main $@
