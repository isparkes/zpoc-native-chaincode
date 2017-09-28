#Init Loyalty chaincode

Create a zip file of the contents of the repository, without git folders: zip -r loyalty.zip *.go vendor mock testdata

Chaincode name: loyalty
Chaincode path: github.com/chaincode
Chaincode version: 1 (or what you have)
Init args: {'admin':'Admin@peer-org0.blockchain-factory.ch'}


#Create Actors

Function: createActors
Transaction type: transaction
Args:
[{'name': 'shop1', 'role': 'shop'}, {'name': 'customer1', 'role': 'customer'}, {'name': 'customer2', 'role': 'customer'}, {'name': 'lbutler', 'role': 'customer'}, {'name': 'wburns', 'role': 'customer'},{'name': 'bank1', 'role': 'bank'}, {'name': 'bank2', 'role': 'bank'}]

#provide asset to a customer (for test)

change the user in Postman using enrollUser as 'bank1' (or 2)
Function: provideAsset
Transaction type: transaction
Args: {'receiver': 'customer1', 'value': 1000}

