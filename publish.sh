#!/bin/bash

source ".env"

assert(){
    if [[ -z "$2" ]]; then
        echo "no $1"
        exit 1
    fi
}

assert CF_ACCOUNT_ID "$CF_ACCOUNT_ID"
assert CF_API_TOKEN "$CF_API_TOKEN"
assert CF_ZONE_ID "$CF_ZONE_ID"
assert CF_ROUTE "$CF_ROUTE"

# hide
wrangler publish >/dev/null 2>&1