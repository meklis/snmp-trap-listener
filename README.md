# SNMP trap listener
Listens for SNMP traps on the specified interface and sends the received packets to a Redis channel in JSON format.

## Build 
**Requires Go 1.19+**
```shell
bash ./build.sh 
```

## Configuration
You can redefine parameters by specifying them through environment variables.   
Alternatively, you can launch the compiled utility with the argument -config specifying the path to the configuration file.   

```yaml 
#Logger configuration
logger:
  console:
    enabled: true
    enable_color: false
    log_level: ${LOG_LEVEL:debug}

listen:
  address: ${LISTEN_ADDRESS:0.0.0.0:162}
  community: ${LISTEN_COMMUNITY}

redis:
  enabled: true
  address: ${REDIS_ADDRESS:127.0.0.1:6379}
  password: ${REDIS_PASSWORD}
  database: ${REDIS_DB:0}
  channel: ${REDIS_CHANNEL:snmptrap}

```

## Message format 
```json 
{
  "host": "127.0.0.1",
  "version": "2c",
  "community": "public",
  "object": ".1.3.6.1.4.1.12345.1",
  "timeticks": 4708949,
  "data": {
    ".1.3.6.1.4.1.12345.1.1": {
      "hex": "54:65:73:74:20:54:72:61:70",
      "type": "OctetString",
      "value": "Test Trap"
    },
    ".1.3.6.1.4.1.12345.1.2": {
      "hex": "54:65:73:74:20:54:72:61:70:32",
      "type": "OctetString",
      "value": "Test Trap2"
    }
  }
}
```