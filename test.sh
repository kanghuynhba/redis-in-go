#!/usr/bin/env bash

TESTER_BIN="/Users/khang/Work/projects/redis-tester/dist/main.out"
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Ensure tester binary exists
if [ ! -f "$TESTER_BIN" ]; then
    echo -e "${YELLOW}Building redis-tester binary...${NC}"
    (cd /Users/khang/Work/projects/redis-tester && go build -o dist/main.out ./cmd/tester)
fi

# All test cases defined in redis-tester
ALL_STAGES=(
    # --- Base Stages (1-7) ---
    "1|jm1|Base: Bind to a port|base"
    "2|rg2|Base: Respond to PING|base"
    "3|wy1|Base: Respond to multiple PINGs|base"
    "4|zu2|Base: Handle concurrent clients|base"
    "5|qq0|Base: Implement the ECHO command|base"
    "6|la7|Base: Implement the SET & GET commands|base"
    "7|yz1|Base: Key Expiry (PX)|base"

    # --- RDB Persistence (8-13) ---
    "8|zg5|RDB: Read RDB config|rdb"
    "9|jz6|RDB: Read key from RDB file|rdb"
    "10|gc6|RDB: Read string value from RDB file|rdb"
    "11|jw4|RDB: Read multiple keys from RDB file|rdb"
    "12|dq3|RDB: Read multiple string values from RDB file|rdb"
    "13|sm4|RDB: Read value with expiry from RDB file|rdb"

    # --- AOF Persistence (14-23) ---
    "14|uj3|AOF: Config defaults|aof"
    "15|vd9|AOF: Config from flags|aof"
    "16|fm0|AOF: Create AOF directory|aof"
    "17|dw4|AOF: Create append-only file|aof"
    "18|pb9|AOF: Create AOF manifest file|aof"
    "19|dc8|AOF: Write single command|aof"
    "20|fi1|AOF: Write multiple commands|aof"
    "21|ep6|AOF: Filter commands before write|aof"
    "22|xz2|AOF: Replay single command|aof"
    "23|kn2|AOF: Replay multiple commands|aof"

    # --- Replication (24-41) ---
    "24|bw1|Replication: Bind to custom port|repl"
    "25|ye5|Replication: INFO replication (role:master)|repl"
    "26|hc6|Replication: INFO replication (role:replica)|repl"
    "27|xc1|Replication: Replication ID & offset|repl"
    "28|gl7|Replication: Replica sends PING|repl"
    "29|eh4|Replication: Replica sends REPLCONF|repl"
    "30|ju6|Replication: Replica sends PSYNC|repl"
    "31|fj0|Replication: Master handles REPLCONF|repl"
    "32|vm3|Replication: Master handles PSYNC|repl"
    "33|cf8|Replication: Master handles PSYNC (RDB transfer)|repl"
    "34|zn8|Replication: Master command propagation|repl"
    "35|hd5|Replication: Multiple replicas|repl"
    "36|yg4|Replication: Command processing on replica|repl"
    "37|xv6|Replication: GETACK with offset 0|repl"
    "38|yd3|Replication: GETACK with non-zero offset|repl"
    "39|my8|Replication: WAIT with 0 replicas|repl"
    "40|tu8|Replication: WAIT with 0 offset|repl"
    "41|na2|Replication: WAIT command|repl"

    # --- Streams (42-54) ---
    "42|cc3|Streams: TYPE command (stream)|streams"
    "43|cf6|Streams: XADD command|streams"
    "44|hq8|Streams: Validate stream ID|streams"
    "45|yh3|Streams: Partially auto-generated ID|streams"
    "46|xu6|Streams: Fully auto-generated ID|streams"
    "47|zx1|Streams: XRANGE command|streams"
    "48|yp1|Streams: XRANGE minimum ID|streams"
    "49|fs1|Streams: XRANGE maximum ID|streams"
    "50|um0|Streams: XREAD command|streams"
    "51|ru9|Streams: XREAD multiple streams|streams"
    "52|bs1|Streams: XREAD with BLOCK|streams"
    "53|hw1|Streams: XREAD with BLOCK (no timeout)|streams"
    "54|xu1|Streams: XREAD with BLOCK (maximum ID)|streams"

    # --- Transactions (55-65) ---
    "55|si4|Transactions: INCR command (basic)|tx"
    "56|lz8|Transactions: INCR command (missing key)|tx"
    "57|mk1|Transactions: INCR command (non-integer)|tx"
    "58|pn0|Transactions: MULTI command|tx"
    "59|lo4|Transactions: EXEC command|tx"
    "60|we1|Transactions: EXEC without MULTI|tx"
    "61|rs9|Transactions: Queue commands in transaction|tx"
    "62|fy6|Transactions: Execute queued commands|tx"
    "63|rl9|Transactions: DISCARD command|tx"
    "64|sg9|Transactions: Errors inside transactions|tx"
    "65|jf8|Transactions: Concurrent transactions|tx"

    # --- Optimistic Locking (66-73) ---
    "66|jb7|Locking: WATCH command|watch"
    "67|jq9|Locking: WATCH inside transaction|watch"
    "68|mh8|Locking: Key modification aborts transaction|watch"
    "69|fp0|Locking: WATCH multiple keys|watch"
    "70|uo9|Locking: WATCH missing key|watch"
    "71|bn1|Locking: UNWATCH command|watch"
    "72|fn4|Locking: UNWATCH on EXEC|watch"
    "73|hq1|Locking: UNWATCH on DISCARD|watch"

    # --- Lists (74-84) ---
    "74|mh6|Lists: RPUSH (single element)|lists"
    "75|tn7|Lists: RPUSH (multiple elements)|lists"
    "76|lx4|Lists: RPUSH (existing list)|lists"
    "77|sf6|Lists: LRANGE (positive indexes)|lists"
    "78|ri1|Lists: LRANGE (negative indexes)|lists"
    "79|gu5|Lists: LPUSH command|lists"
    "80|fv6|Lists: LLEN command|lists"
    "81|ef1|Lists: LPOP (single element)|lists"
    "82|jp1|Lists: LPOP (multiple elements)|lists"
    "83|ec3|Lists: BLPOP without timeout|lists"
    "84|xj7|Lists: BLPOP with timeout|lists"

    # --- Pub/Sub (85-91) ---
    "85|mx3|PubSub: SUBSCRIBE command|pubsub"
    "86|zc8|PubSub: SUBSCRIBE multiple channels|pubsub"
    "87|aw8|PubSub: SUBSCRIBE to existing channel|pubsub"
    "88|lf1|PubSub: Multiple SUBSCRIBE commands|pubsub"
    "89|hf2|PubSub: PUBLISH to single subscriber|pubsub"
    "90|dn4|PubSub: PUBLISH to multiple subscribers|pubsub"
    "91|ze9|PubSub: UNSUBSCRIBE command|pubsub"

    # --- Sorted Sets (92-99) ---
    "92|ct1|ZSet: ZADD (single member)|zset"
    "93|hf1|ZSet: ZADD (multiple members)|zset"
    "94|lg6|ZSet: ZRANK command|zset"
    "95|ic1|ZSet: ZRANGE (positive indexes)|zset"
    "96|bj4|ZSet: ZRANGE (negative indexes)|zset"
    "97|kn4|ZSet: ZCARD command|zset"
    "98|gd7|ZSet: ZSCORE command|zset"
    "99|sq7|ZSet: ZREM command|zset"

    # --- Bitmaps (100-108) ---
    "100|bq9|Bitmaps: SETBIT (single bit)|bitmaps"
    "101|qj1|Bitmaps: GETBIT (single bit)|bitmaps"
    "102|yj2|Bitmaps: Read string as bits|bitmaps"
    "103|pk5|Bitmaps: Read bits as string|bitmaps"
    "104|yf6|Bitmaps: SETBIT growing bitmap|bitmaps"
    "105|nx3|Bitmaps: BITCOUNT command|bitmaps"
    "106|hv4|Bitmaps: BITOP AND|bitmaps"
    "107|dk2|Bitmaps: BITOP AND (diff lengths)|bitmaps"
    "108|fr8|Bitmaps: BITOP OR|bitmaps"

    # --- Geospatial (109-116) ---
    "109|zt4|Geo: GEOADD command|geo"
    "110|ck3|Geo: GEOADD (validate coordinates)|geo"
    "111|tn5|Geo: GEOADD (store location)|geo"
    "112|cr3|Geo: GEOADD (calculate score)|geo"
    "113|xg4|Geo: GEOPOS command|geo"
    "114|hb5|Geo: GEOPOS (decode coordinates)|geo"
    "115|ek6|Geo: GEODIST command|geo"
    "116|rm9|Geo: GEOSEARCH command|geo"

    # --- Auth & ACL (117-124) ---
    "117|jn4|ACL: WHOAMI command|acl"
    "118|gx8|ACL: GETUSER command|acl"
    "119|ql6|ACL: GETUSER (nopass flag)|acl"
    "120|pl7|ACL: GETUSER (passwords)|acl"
    "121|uv9|ACL: SETUSER (passwords)|acl"
    "122|hz3|ACL: AUTH command response|acl"
    "123|nm2|ACL: Default user authentication|acl"
    "124|ws7|ACL: AUTH command authentication|acl"
)

run_single_stage() {
    local num="$1"
    local slug="$2"
    local title="$3"
    local silent="$4"

    local json="[{\"slug\":\"$slug\",\"tester_log_prefix\":\"stage-$num\",\"title\":\"$title\"}]"
    local tmp_output
    tmp_output=$(mktemp)

    if [ "$silent" = "true" ]; then
        if CODECRAFTERS_REPOSITORY_DIR="$REPO_DIR" CODECRAFTERS_TEST_CASES_JSON="$json" "$TESTER_BIN" > "$tmp_output" 2>&1; then
            echo -e "  ${GREEN}✔${NC}  ${BOLD}Stage #$num${NC} [${CYAN}$slug${NC}]: $title ${GREEN}(Passed)${NC}"
            rm -f "$tmp_output"
            return 0
        else
            echo -e "\n  ${RED}✖${NC}  ${BOLD}Stage #$num${NC} [${CYAN}$slug${NC}]: $title ${RED}(Failed!)${NC}\n"
            echo -e "${RED}─────────────────── Test Failure Output ───────────────────${NC}"
            cat "$tmp_output"
            echo -e "${RED}───────────────────────────────────────────────────────────${NC}\n"
            rm -f "$tmp_output"
            return 1
        fi
    else
        CODECRAFTERS_REPOSITORY_DIR="$REPO_DIR" CODECRAFTERS_TEST_CASES_JSON="$json" "$TESTER_BIN"
    fi
}

run_group() {
    local target_group="$1"
    local group_label="$2"

    echo -e "${BLUE}${BOLD}=== Running $group_label Tests ===${NC}\n"
    local all_passed=true

    for item in "${ALL_STAGES[@]}"; do
        IFS="|" read -r num slug title group <<< "$item"
        if [ "$target_group" = "all" ] || [ "$group" = "$target_group" ]; then
            if ! run_single_stage "$num" "$slug" "$title" "true"; then
                all_passed=false
                echo -e "${YELLOW}👉 Fix the error above in Stage #$num ($title) and run again!${NC}\n"
                exit 1
            fi
        fi
    done

    if [ "$all_passed" = "true" ]; then
        echo -e "\n${GREEN}${BOLD}🎉 ALL $group_label TESTS PASSED! 🎉${NC}\n"
    fi
}

TARGET="${1:-}"

case "$TARGET" in
    "")
        # Default: run base stages waterfall (1-7)
        run_group "base" "Base Stages (1-7)"
        ;;
    all)
        run_group "all" "ALL (1-124)"
        ;;
    base)
        run_group "base" "Base Stages (1-7)"
        ;;
    rdb)
        run_group "rdb" "RDB Persistence"
        ;;
    aof)
        run_group "aof" "AOF Persistence"
        ;;
    repl|replication)
        run_group "repl" "Replication"
        ;;
    streams|stream)
        run_group "streams" "Streams"
        ;;
    tx|transactions)
        run_group "tx" "Transactions"
        ;;
    watch|locking)
        run_group "watch" "Optimistic Locking"
        ;;
    lists|list)
        run_group "lists" "Lists"
        ;;
    pubsub)
        run_group "pubsub" "Pub/Sub"
        ;;
    zset|sorted_sets)
        run_group "zset" "Sorted Sets"
        ;;
    bitmaps|bit)
        run_group "bitmaps" "Bitmaps"
        ;;
    geo|geospatial)
        run_group "geo" "Geospatial"
        ;;
    acl|auth)
        run_group "acl" "Auth & ACL"
        ;;
    list-all|stages)
        echo -e "${BOLD}Available Stages:${NC}"
        for item in "${ALL_STAGES[@]}"; do
            IFS="|" read -r num slug title group <<< "$item"
            printf "  ${YELLOW}%3d${NC} | ${CYAN}%-6s${NC} | ${MAGENTA}%-10s${NC} | %s\n" "$num" "$slug" "[$group]" "$title"
        done
        ;;
    *)
        # Search by number or slug
        found=false
        for item in "${ALL_STAGES[@]}"; do
            IFS="|" read -r num slug title group <<< "$item"
            if [ "$TARGET" = "$num" ] || [ "$TARGET" = "$slug" ]; then
                found=true
                run_single_stage "$num" "$slug" "$title" "false"
                break
            fi
        done

        if [ "$found" = "false" ]; then
            echo -e "${RED}Unknown stage or category: '$TARGET'${NC}\n"
            echo -e "${BOLD}Usage:${NC}"
            echo -e "  ./test.sh                 - Run Base stages waterfall (1-7)"
            echo -e "  ./test.sh all             - Run ALL 124 stages across all extensions"
            echo -e "  ./test.sh <number|slug>   - Run a specific stage (e.g. ./test.sh 1, ./test.sh yz1)"
            echo -e "  ./test.sh <group>         - Run group (base, rdb, aof, repl, streams, tx, watch, lists, pubsub, zset, bitmaps, geo, acl)"
            echo -e "  ./test.sh list-all        - List all 124 available stages"
            exit 1
        fi
        ;;
esac
