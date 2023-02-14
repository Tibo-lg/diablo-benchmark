#/bin/sh

CONTRACT_FOLDER="${CONTRACT_FOLDER:-solidity-contracts}"

for f in $(ls ./${CONTRACT_FOLDER}/); do
	for c in $(ls ./${CONTRACT_FOLDER}/${f}/*.sol); do
		solc --combined-json abi,bin,hashes ${c} -o ${CONTRACT_FOLDER}/${f}/ --overwrite
	done
done
