
# simple environment setup and then runs the go test


export RHT_PW=BunnyHop				# replace with your password
export RHT_USER=scott				# replace with your user name
export RHT_HOST=$RMQ_HOST			# replace with IP/host where rmq is running

export RHT_PAUSE=0
export RHT_DEL_EX=1

export RHT_EXTYPE="fanout+!du>rht_queue+!du+ad+ex"
export RHT_EXTYPE="fanout+!du>random+!du+ad+ex"

go test
