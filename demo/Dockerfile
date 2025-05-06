FROM postgres:15

RUN apt-get update && apt-get install -y \
    git \
    build-essential \
    postgresql-server-dev-15 \
 && git clone https://github.com/eulerto/wal2json.git \
 && cd wal2json \
 && make && make install \
 && cd .. && rm -rf wal2json
