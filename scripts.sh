# add 100ms delay  in  localhost
sudo tc qdisc add dev lo root netem delay 100ms

# remove 100ms delay in localhost 
sudo tc qdisc del dev lo root netem

