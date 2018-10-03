args = process.argv

// first arg is node
// second arg is relative filepath of this file

console.log("args "+ args)
if (args[2] != "-c") {
  console.log("Expected flag -c to be provided");
  process.exit(1);
}

configFile = args[3]
if (configFile == "") {
  console.log("Deployer Config has not been provided");
  process.exit(1);
}
var fs = require("fs");
configContent = fs.readFileSync(configFile);
var config = JSON.parse(configContent);

if (config["provider"] != "http://127.0.0.1:1234") {
  console.log("Incorrect provider. Expected http://127.0.0.1:1234, received " + config["provider"]);
  process.exit(1);
}

if (config["password"] != "") {
  console.log("Password should be empty");
  process.exit(1);
}

contractArgs = config["args"]
if (contractArgs.length != 2) {
  console.log("Args were not passed in untouched");
  process.exit(1);
}

if (contractArgs[0] != "sample-arg-1") {
  console.log("Arg1 is in incorrect. Expected sample-arg-1, received " + contractArgs[0])
  process.exit(1);
}

if (contractArgs[1] != "sample-arg-2") {
  console.log("Arg2 is in incorrect. Expected sample-arg-2, received " + contractArgs[1])
  process.exit(1);
}


if (args[4] != "-o") {
  console.log("Expected flag -o to be provided");
  process.exit(1);
}

outputFile = args[5]
if (outputFile == "") {
  console.log("Output file name has not been provided");
  process.exit(1);
}

// hardcoded fake contract path for test
if (args[6] != "path-to-contract") {
  console.log("Incorrect contract path was passed in. Expected path-to-contract, received " + args[6]);
  process.exit(1);
}

var fs = require('fs');
var result = {};
result["address"] = "sample-account";
result["abi"] = "sample-abi";
result["contract_address"] = "sample-address";
result["gas_price"] = "0";
result["transaction_hash"] = "sample-tx-hash";

fs.writeFile(outputFile, JSON.stringify(result, null, 4), (err) => {
  if (err) {
    console.error("error writing file" + err);
    process.exit(1)
  };
});
