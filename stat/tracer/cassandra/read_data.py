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

def readRows():
    cluster = Cluster(contact_points=['127.0.0.1'],port=9142)
    session = cluster.connect()

    log.info("setting keyspace...")
    session.set_keyspace(KEYSPACE)

    rows = session.execute("SELECT * FROM mytable")
    log.info("key\tcol1\tcol2")
    log.info("---------\t----\t----")

    count=0
    for row in rows:
        if(count%100==0):
            log.info('\t'.join(row))
        count=count+1;

    log.info("Total")
    log.info("-----")
    log.info("rows %d" %(count))

readRows()
