#!/bin/sh -e

: ${ALERTMANAGER:="http://localhost:9093"}
: ${PORT0:="8080"}
: ${LOG_LEVEL:="warn"}
: ${LOG_FORMATTER:="text"}
: ${PRUNE_AGE:="1m"}
: ${FORGET_AGE:="1h"}
: ${PRUNE_INTERVAL:="2m"}
: ${PUBLISH_INTERVAL:="1m"}
: ${PUBLISH_MINIMUM:="5s"}

exec promsaint \
    -alertmanager $ALERTMANAGER \
    -listen ":$PORT0" \
    -log.level $LOG_LEVEL \
    -log.format $LOG_FORMATTER \
    -pruneage $PRUNE_AGE \
    -forgetage $FORGET_AGE \
    -pruneinterval $PRUNE_INTERVAL \
    -publishinterval $PUBLISH_INTERVAL \
    -publishminimum $PUBLISH_MINIMUM
