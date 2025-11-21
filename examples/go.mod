module github.com/spdeepak/flowtracker/examples

go 1.24

require (
	github.com/confluentinc/confluent-kafka-go v1.9.2
	github.com/spdeepak/flowtracker v0.0.3
	github.com/spdeepak/flowtracker/confluent-kafka v0.0.1
)

replace (
	github.com/spdeepak/flowtracker => ../
	github.com/spdeepak/flowtracker/confluent-kafka => ../addons/confluent-kafka/
)
