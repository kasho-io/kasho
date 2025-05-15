import time
import json
import os
import psycopg2
from psycopg2.extras import LogicalReplicationConnection
from psycopg2.extras import RealDictCursor
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

REPLICA_CONN = {
    'dbname': os.getenv('REPLICA_DB'),
    'user': os.getenv('REPLICA_USER'),
    'password': os.getenv('REPLICA_PASSWORD'),
    'host': os.getenv('REPLICA_HOST'),
    'port': int(os.getenv('REPLICA_PORT'))
}

PRIMARY_CONN = {
    'dbname': os.getenv('PRIMARY_DB'),
    'user': os.getenv('PRIMARY_USER'),
    'password': os.getenv('PRIMARY_PASSWORD'),
    'host': os.getenv('PRIMARY_HOST'),
    'port': int(os.getenv('PRIMARY_PORT'))
}

REPLICATION_SLOT = 'translicate_slot'
REPLICATION_PUBLICATION = 'translicate_pub'

DDL_LOG_QUERY = """
    SELECT * FROM translicate_ddl_log
    WHERE lsn > %s::pg_lsn
    ORDER BY lsn
"""

BUFFERED_DMLS = []

SEQ_QUERY = """
SELECT
  n.nspname AS schema,
  t.relname AS table,
  a.attname AS column,
  s.relname AS sequence
FROM pg_class s
JOIN pg_depend d ON d.objid = s.oid AND d.deptype = 'a'
JOIN pg_class t ON d.refobjid = t.oid
JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = d.refobjsubid
JOIN pg_namespace n ON n.oid = t.relnamespace
WHERE s.relkind = 'S';
"""

def lsn_to_int(lsn_str):
    if lsn_str == '0/0':
        return 0
    high, low = lsn_str.split('/')
    return (int(high, 16) << 32) + int(low, 16)

def int_to_lsn(lsn_int):
    if lsn_int == 0:
        return '0/0'
    high = lsn_int >> 32
    low = lsn_int & 0xFFFFFFFF
    return f"{high:x}/{low:x}"

def apply_ddl(ddl, ddl_lsn, replica_cursor):
    print(f"== APPLYING DDL @ {ddl_lsn} ==")
    print(ddl)
    try:
        replica_cursor.execute(ddl)
        return ddl_lsn
    except Exception as e:
        print(f"Error applying DDL: {e}")
        return None

def apply_dml(change, replica_cursor):
    table = change['table']
    columns = change['columnnames']
    values = change['columnvalues']
    operation = change['kind']
    
    try:
        if operation == 'insert':
            sql = f"INSERT INTO {table} ({', '.join(columns)}) VALUES ({', '.join(['%s']*len(values))})"
            print(f"== APPLYING INSERT: {sql} {values}")
            replica_cursor.execute(sql, values)
        elif operation == 'update':
            # For updates, we need both old and new values
            old_values = change['oldkeys']['keyvalues']
            set_clause = ', '.join(f"{col} = %s" for col in columns)
            where_clause = ' AND '.join(f"{col} = %s" for col in change['oldkeys']['keynames'])
            sql = f"UPDATE {table} SET {set_clause} WHERE {where_clause}"
            print(f"== APPLYING UPDATE: {sql} {values + old_values}")
            replica_cursor.execute(sql, values + old_values)
        elif operation == 'delete':
            # For deletes, we use the old values to identify the row
            old_values = change['oldkeys']['keyvalues']
            where_clause = ' AND '.join(f"{col} = %s" for col in change['oldkeys']['keynames'])
            sql = f"DELETE FROM {table} WHERE {where_clause}"
            print(f"== APPLYING DELETE: {sql} {old_values}")
            replica_cursor.execute(sql, old_values)
        else:
            print(f"Unsupported operation type: {operation}")
    except Exception as e:
        print(f"Error applying DML: {e}")

def apply_sequence_change(change, replica_cursor):
    sequence = change['sequence']
    value = change['value']
    try:
        sql = f"SELECT setval('{sequence}', {value}, true)"
        print(f"== APPLYING SEQUENCE CHANGE: {sql}")
        replica_cursor.execute(sql)
    except Exception as e:
        print(f"Error applying sequence change: {e}")

def sync_sequences(replica_cursor):
    replica_cursor.execute(SEQ_QUERY)
    for schema, table, column, sequence in replica_cursor.fetchall():
        full_table = f"{schema}.{table}"
        full_seq = f"{schema}.{sequence}"
        print(f"Syncing {full_seq} to MAX({column}) in {full_table}")
        replica_cursor.execute(f"""
            SELECT setval(
                %s,
                COALESCE((SELECT MAX({column}) FROM {full_table}), 1)
            );
        """, (full_seq,))

def poll_and_apply_ddls(primary_cursor, replica_cursor, last_lsn_str):
    print(f"Polling for DDLs after LSN {last_lsn_str}")
    print(f"Executing query: {DDL_LOG_QUERY} with param: {last_lsn_str}")
    primary_cursor.execute(DDL_LOG_QUERY, (last_lsn_str,))
    ddls = primary_cursor.fetchall()
    print(f"Found {len(ddls)} DDLs to apply")
    applied_lsn_str = last_lsn_str
    
    for ddl_row in ddls:
        lsn = ddl_row['lsn']
        ddl = ddl_row['ddl']
        # Skip DDLs related to translicate_ddl_log or publications
        if 'translicate_ddl_log' in ddl.lower() or 'publication' in ddl.lower():
            print(f"Skipping DDL at LSN {lsn} (translicate_ddl_log or publication related)")
            applied_lsn_str = lsn
            continue
            
        print(f"Attempting to apply DDL at LSN {lsn}: {ddl}")
        try:
            replica_cursor.execute(ddl)
            print(f"Successfully applied DDL at LSN {lsn}")
            applied_lsn_str = lsn
        except Exception as e:
            print(f"Error applying DDL at LSN {lsn}: {e}")
            # Don't continue if a DDL fails
            break
    
    return applied_lsn_str

def main():
    applied_ddl_lsn_str = None
    applied_ddl_lsn_int = None

    # Connect to primary for polling ddl_log
    ddl_conn = psycopg2.connect(**PRIMARY_CONN)
    ddl_cursor = ddl_conn.cursor(cursor_factory=RealDictCursor)

    # Connect to replica for applying DDLs and DMLs
    replica_conn = psycopg2.connect(**REPLICA_CONN)
    replica_conn.autocommit = True
    replica_cursor = replica_conn.cursor()

    # Connect to logical replication slot
    stream_conn = psycopg2.connect(connection_factory=LogicalReplicationConnection, **PRIMARY_CONN)
    stream_cursor = stream_conn.cursor()
    stream_cursor.start_replication(slot_name=REPLICATION_SLOT, decode=True)

    # Start from the beginning of the DDL log
    applied_ddl_lsn_str = '0/0'
    applied_ddl_lsn_int = lsn_to_int(applied_ddl_lsn_str)

    def handle_message(msg):
        nonlocal applied_ddl_lsn_str, applied_ddl_lsn_int
        print("== WAL2JSON EVENT ==")
        event = json.loads(msg.payload)
        print(json.dumps(event, indent=2))

        print(f"Current LSN: {msg.data_start}")
        print(f"Applied DDL LSN: {applied_ddl_lsn_str} (int: {applied_ddl_lsn_int})")

        wal_lsn_str = int_to_lsn(msg.data_start)
        print(f"WAL LSN as string: {wal_lsn_str}")

        # First, poll and apply any pending DDLs
        applied_ddl_lsn_str = poll_and_apply_ddls(ddl_cursor, replica_cursor, applied_ddl_lsn_str)
        applied_ddl_lsn_int = lsn_to_int(applied_ddl_lsn_str)
        print(f"After polling DDLs, Applied DDL LSN: {applied_ddl_lsn_str} (int: {applied_ddl_lsn_int})")

        changes = event.get("change", [])
        print(f"Number of changes in event: {len(changes)}")
        
        # Process all DMLs in LSN order, skipping translicate_ddl_log
        for change in changes:
            if change.get('table') == 'translicate_ddl_log':
                print("Skipping translicate_ddl_log change...")
                continue
            
            lsn = msg.data_start
            print(f"Processing DML for table: {change.get('table')} with LSN: {lsn}")
            
            if lsn > applied_ddl_lsn_int:
                print(f"Buffering DML due to LSN {lsn} > {applied_ddl_lsn_int}")
                BUFFERED_DMLS.append((lsn, change))
            else:
                print(f"Applying DML immediately")
                apply_dml(change, replica_cursor)

        # After processing all changes, try to apply any remaining buffered DMLs
        if BUFFERED_DMLS:
            print("Checking remaining buffered DMLs...")
            still_buffered = []
            for lsn, change in BUFFERED_DMLS:
                if lsn <= msg.data_start:  # Compare against current WAL LSN
                    print(f"Applying remaining buffered DML with LSN: {lsn}")
                    apply_dml(change, replica_cursor)
                else:
                    still_buffered.append((lsn, change))
            BUFFERED_DMLS[:] = still_buffered
            print(f"Final number of buffered DMLs: {len(BUFFERED_DMLS)}")

        # Sync sequences after processing all DMLs
        if changes:
            print("Syncing sequences after DML changes")
            sync_sequences(replica_cursor)

        msg.cursor.send_feedback(flush_lsn=msg.data_start)

    stream_cursor.consume_stream(handle_message)

if __name__ == '__main__':
    main()
