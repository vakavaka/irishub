echo 12345678 | iriscli --home ibc-iris/n0/iriscli tx ibcmockbank transfer \
    --src-port bank --src-channel chann-to-gaia \
    --denom uiris --amount 1000000 \
    --receiver $(gaiacli --home ibc-gaia/n0/gaiacli keys show n0 | jq -r '.address') \
    --source true \
    --from n0 -y -o json \
    --broadcast-mode=block
