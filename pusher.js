const CLI = require('cli-flags')
const {flags, args} = CLI.parse({
  flags: {
    'config': CLI.flags.string({char: 'c'}),
    'output': CLI.flags.string({char: 'o'})
  },
  args: [
    {name: 'contract', required: true}
  ]
})

var fs = require("fs");
configContent = fs.readFileSync(flags.config);
var outputFile = flags.output
var config = JSON.parse(configContent);

const provider = config.provider
const password = config.password
const contractArgs = config.args
const nodeType = config.type
const contractPath = args.contract

result = {}
const Web3 = require('web3')
if (typeof web3 !== 'undefined') {
  web3 = new Web3(web3.currentProvider);
} else {
  // set the provider you want from Web3.providers
  web3 = new Web3(new Web3.providers.HttpProvider(provider));
}

if (provider == "") {
    console.log("no provider is provided")
    process.exit()
}

var fs = require('fs');
if (!fs.existsSync(contractPath)) {
    console.log("contract path incorrect: " + contractPath)
    process.exit()
}

fs.readFile(contractPath, {encoding: 'utf-8'}, function(err,data){
    if (!err) {
      var solc = require('solc')
      var output = solc.compile(data, 1)
      for(var contractName in output.contracts){
        runDeployment(output.contracts[contractName])
      }
    } else {
      console.log(err);
      process.exit()
    }
})


function convertArray(argument){
  var convertedArgs = []
  if (argument == null) {
    return convertedArgs
  }

  argument.forEach(function(arg){
    if (Array.isArray(arg)){
      convertedArgs.push(convertArray(arg));
    } else {
      convertedArgs.push(web3.utils.asciiToHex(arg)) ;
    }
  });
  return convertedArgs
}

const constructorArgs = convertArray(contractArgs)

function runDeployment(contract){
  web3.eth.getAccounts()
  .then(addresses => {
    address = addresses[0]
    result["address"] = address
    return (nodeType == "ethereum") ? web3.eth.personal.unlockAccount(address, '') : address
  })
  .then(() => {
    var contractAbi = contract.interface
    result["abi"] = contractAbi
    var contractInterface = new web3.eth.Contract(JSON.parse(contractAbi))
    var compiled = "0x" + contract.bytecode
    deploy = contractInterface.deploy({
      data: compiled,
      arguments: constructorArgs
    });
    return deploy.estimateGas({from: address})
  })
  .then(gasEstimate => {
    return deploy.send({
      from: address,
      gas: gasEstimate,
      gasLimit: 4000000
    })
    .once('transactionHash', hash => {
      result['transaction_hash'] = hash
      web3.eth.getTransaction(hash)
      .then(transaction => result['gas_price'] = transaction.gasPrice);
      web3.eth.getTransactionReceipt(hash)
      .then(transactionReceipt => result['contract_address'] = transactionReceipt.contractAddress); //console.log("contract address: " + transactionReceipt.contractAddress));
    })
  })
  .then(() => {
    fs.writeFile(outputFile, JSON.stringify(result, null, 4), (err) => {
        if (err) {
            console.error(err);
            return;
        };
        console.log("File has been created");
    });
  })
  .catch(err => console.log(err.stack))
}
