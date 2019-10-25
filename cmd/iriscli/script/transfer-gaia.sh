echo 12345678 | gaiacli --home ibc-gaia/n0/gaiacli tx ibcmockbank transfer \
    --src-port bank --src-channel chann-to-iris \
    --denom uatom --amount 1000000 \
    --receiver $(iriscli --home ibc-iris/n0/iriscli keys show n0 | jq -r '.address') \
    --source true \
    --from n0 -y -o json \
    --broadcast-mode=block
