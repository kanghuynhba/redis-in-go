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

# All test cases matching CodeCrafters official stage order (1-124)
ALL_STAGES=(
    # --- 1. Base Stages (1-7) ---
    "1|jm1|Bind to a port|base"
    "2|rg2|Respond to PING|base"
    "3|wy1|Respond to multiple PINGs|base"
    "4|zu2|Handle concurrent clients|base"
    "5|qq0|Implement the ECHO command|base"
    "6|la7|Implement the SET & GET commands|base"
    "7|yz1|Expiry|base"

    # --- 2. Lists (8-18) ---
    "8|mh6|Create a list|lists"
    "9|tn7|Append an element|lists"
    "10|lx4|Append multiple elements|lists"
    "11|sf6|List elements (positive indexes)|lists"
    "12|ri1|List elements (negative indexes)|lists"
    "13|gu5|Prepend elements|lists"
    "14|fv6|Query list length|lists"
    "15|ef1|Remove an element|lists"
    "16|jp1|Remove multiple elements|lists"
    "17|ec3|Blocking retrieval|lists"
    "18|xj7|Blocking retrieval with timeout|lists"

    # --- 3. Streams (19-31) ---
    "19|cc3|The TYPE command|streams"
    "20|cf6|Create a stream|streams"
    "21|hq8|Validating entry IDs|streams"
    "22|yh3|Partially auto-generated IDs|streams"
    "23|xu6|Fully auto-generated IDs|streams"
    "24|zx1|Query entries from stream|streams"
    "25|yp1|Query with -|streams"
    "26|fs1|Query with +|streams"
    "27|um0|Query single stream using XREAD|streams"
    "28|ru9|Query multiple streams using XREAD|streams"
    "29|bs1|Blocking reads|streams"
    "30|hw1|Blocking reads without timeout|streams"
    "31|xu1|Blocking reads using $|streams"

    # --- 4. Transactions (32-42) ---
    "32|si4|The INCR command (1/3)|tx"
    "33|lz8|The INCR command (2/3)|tx"
    "34|mk1|The INCR command (3/3)|tx"
    "35|pn0|The MULTI command|tx"
    "36|lo4|The EXEC command|tx"
    "37|we1|Empty transaction|tx"
    "38|rs9|Queueing commands|tx"
    "39|fy6|Executing a transaction|tx"
    "40|rl9|The DISCARD command|tx"
    "41|sg9|Failures within transactions|tx"
    "42|jf8|Multiple transactions|tx"

    # --- 5. Optimistic Locking (43-50) ---
    "43|jb7|The WATCH command|watch"
    "44|jq9|WATCH inside transaction|watch"
    "45|mh8|Tracking key modifications|watch"
    "46|fp0|Watching multiple keys|watch"
    "47|uo9|Watching missing keys|watch"
    "48|bn1|The UNWATCH command|watch"
    "49|fn4|Unwatch on EXEC|watch"
    "50|hq1|Unwatch on DISCARD|watch"

    # --- 6. Replication (51-68) ---
    "51|bw1|Configure listening port|repl"
    "52|ye5|The INFO command|repl"
    "53|hc6|The INFO command on a replica|repl"
    "54|xc1|Initial replication ID and offset|repl"
    "55|gl7|Send handshake (1/3)|repl"
    "56|eh4|Send handshake (2/3)|repl"
    "57|ju6|Send handshake (3/3)|repl"
    "58|fj0|Receive handshake (1/2)|repl"
    "59|vm3|Receive handshake (2/2)|repl"
    "60|cf8|Empty RDB transfer|repl"
    "61|zn8|Single-replica propagation|repl"
    "62|hd5|Multi-replica propagation|repl"
    "63|yg4|Command processing|repl"
    "64|xv6|ACKs with no commands|repl"
    "65|yd3|ACKs with commands|repl"
    "66|my8|WAIT with no replicas|repl"
    "67|tu8|WAIT with no commands|repl"
    "68|na2|WAIT with multiple commands|repl"

    # --- 7. RDB Persistence (69-74) ---
    "69|zg5|RDB file config|rdb"
    "70|jz6|Read a key|rdb"
    "71|gc6|Read a string value|rdb"
    "72|jw4|Read multiple keys|rdb"
    "73|dq3|Read multiple string values|rdb"
    "74|sm4|Read value with expiry|rdb"

    # --- 8. AOF Persistence (75-84) ---
    "75|uj3|Default AOF options|aof"
    "76|vd9|AOF options from flags|aof"
    "77|fm0|Create append-only directory|aof"
    "78|dw4|Create append-only file|aof"
    "79|pb9|Create manifest file|aof"
    "80|dc8|Write a single command|aof"
    "81|fi1|Write multiple commands|aof"
    "82|ep6|Filter write commands|aof"
    "83|xz2|Replay a single command|aof"
    "84|kn2|Replay multiple commands|aof"

    # --- 9. Pub/Sub (85-91) ---
    "85|mx3|Subscribe to a channel|pubsub"
    "86|zc8|Subscribe to multiple channels|pubsub"
    "87|aw8|Enter subscribed mode|pubsub"
    "88|lf1|PING in subscribed mode|pubsub"
    "89|hf2|Publish a message|pubsub"
    "90|dn4|Deliver messages|pubsub"
    "91|ze9|Unsubscribe|pubsub"

    # --- 10. Sorted Sets (92-99) ---
    "92|ct1|Create a sorted set|zset"
    "93|hf1|Add members|zset"
    "94|lg6|Retrieve member rank|zset"
    "95|ic1|List sorted set members|zset"
    "96|bj4|ZRANGE with negative indexes|zset"
    "97|kn4|Count sorted set members|zset"
    "98|gd7|Retrieve member score|zset"
    "99|sq7|Remove a member|zset"

    # --- 11. Bitmaps (100-108) ---
    "100|bq9|Create a bitmap|bitmaps"
    "101|qj1|Retrieve a bit|bitmaps"
    "102|yj2|Read a string as bits|bitmaps"
    "103|pk5|Read bits as a string|bitmaps"
    "104|yf6|Grow a bitmap|bitmaps"
    "105|nx3|Count set bits|bitmaps"
    "106|hv4|AND two bitmaps|bitmaps"
    "107|dk2|AND bitmaps of different lengths|bitmaps"
    "108|fr8|OR two bitmaps|bitmaps"

    # --- 12. Geospatial Commands (109-116) ---
    "109|zt4|Respond to GEOADD|geo"
    "110|ck3|Validate coordinates|geo"
    "111|tn5|Store a location|geo"
    "112|cr3|Calculate location score|geo"
    "113|xg4|Respond to GEOPOS|geo"
    "114|hb5|Decode coordinates|geo"
    "115|ek6|Calculate distance|geo"
    "116|rm9|Search within radius|geo"

    # --- 13. Authentication (117-124) ---
    "117|jn4|Respond to ACL WHOAMI|acl"
    "118|gx8|Respond to ACL GETUSER|acl"
    "119|ql6|The nopass flag|acl"
    "120|pl7|The passwords property|acl"
    "121|uv9|Setting default user password|acl"
    "122|hz3|The AUTH command|acl"
    "123|nm2|Enforce authentication|acl"
    "124|ws7|Authenticate using AUTH|acl"
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
    lists|list)
        run_group "lists" "Lists (8-18)"
        ;;
    streams|stream)
        run_group "streams" "Streams (19-31)"
        ;;
    tx|transactions)
        run_group "tx" "Transactions (32-42)"
        ;;
    watch|locking)
        run_group "watch" "Optimistic Locking (43-50)"
        ;;
    repl|replication)
        run_group "repl" "Replication (51-68)"
        ;;
    rdb)
        run_group "rdb" "RDB Persistence (69-74)"
        ;;
    aof)
        run_group "aof" "AOF Persistence (75-84)"
        ;;
    pubsub)
        run_group "pubsub" "Pub/Sub (85-91)"
        ;;
    zset|sorted_sets)
        run_group "zset" "Sorted Sets (92-99)"
        ;;
    bitmaps|bit)
        run_group "bitmaps" "Bitmaps (100-108)"
        ;;
    geo|geospatial)
        run_group "geo" "Geospatial (109-116)"
        ;;
    acl|auth)
        run_group "acl" "Authentication (117-124)"
        ;;
    list-all|stages)
        echo -e "${BOLD}Available Stages (CodeCrafters Official Order 1-124):${NC}"
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
            echo -e "  ./test.sh <number|slug>   - Run a specific stage (e.g. ./test.sh 8, ./test.sh mh6)"
            echo -e "  ./test.sh <group>         - Run group (base, lists, streams, tx, watch, repl, rdb, aof, pubsub, zset, bitmaps, geo, acl)"
            echo -e "  ./test.sh list-all        - List all 124 available stages in official order"
            exit 1
        fi
        ;;
esac
