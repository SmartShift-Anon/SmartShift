// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;
//337691 gas
contract Contract8 {
   uint256 public varUint;
   uint64[3] public arrFixedUint64;
   string[] public arrDynamicString;


   struct StructType {
       string varString;
       uint256 varUint;
       uint256[] arrUint;
   }


   StructType[2] public arrStructFixed; // Fixed-size array of structs


   function initialize() public {
       // Initialize variables outside the struct
       varUint = 12345;
       arrFixedUint64 = [1, 2, 3];
       arrDynamicString.push("First");
       arrDynamicString.push("Second");


       // Initialize the fixed-size array of structs
       arrStructFixed[0].varString = "First Struct";
       arrStructFixed[0].varUint = 100;
       arrStructFixed[0].arrUint.push(11);
       arrStructFixed[0].arrUint.push(22);


       arrStructFixed[1].varString = "Second Struct";
       arrStructFixed[1].varUint = 200;
       arrStructFixed[1].arrUint.push(33);
       arrStructFixed[1].arrUint.push(44);
   }
}
