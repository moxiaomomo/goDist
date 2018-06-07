import logging

log = logging.getLogger()
log.setLevel('INFO')
handler = logging.StreamHandler()
handler.setFormatter(logging.Formatter("%(asctime)s [%(levelname)s] %(name)s: %(message)s"))
log.addHandler(handler)

from cassandra import ConsistencyLevel
from cassandra.cluster import Cluster
from cassandra.query import SimpleStatement

KEYSPACE = "mykeyspace"

def insertData(number):
    cluster = Cluster(contact_points=['127.0.0.1'],port=9042)
    session = cluster.connect()

    log.info("setting keyspace...")
    session.set_keyspace(KEYSPACE)

    prepared = session.prepare("""
    INSERT INTO mytable (mykey, col1, col2)
    VALUES (?, ?, ?)
    """)

    for i in range(number):
        if(i%100 == 0):
            log.info("inserting row %d" % i)
        session.execute(prepared.bind(("rec_key_%d" % i, 'aaa', 'bbb')))

insertData(1000)
