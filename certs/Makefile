# Generate CA cert
# cfssl genkey -initca ca-csr.json | cfssljson -bare ca

# Generate proxy cert, signed by ca
# cfssl gencert -ca ./ca.pem -ca-key ./ca-key.pem -hostname=127.0.0.1,proxy ./proxy-csr.json | cfssljson -bare proxy

# Generate promethues client cert, signed by ca
# cfssl gencert -ca ./ca.pem -ca-key ./ca-key.pem -hostname=prometheus ./prometheus-csr.json  | cfssljson -bare prometheus

