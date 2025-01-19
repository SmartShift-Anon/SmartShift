// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;
//135772 gas
contract Contract7 {
   string public varString;
   uint256 public varUint;
   uint64[] public arrDynamicUint64;


   struct StructType {
       string varString;
       uint256 varUint;
   }


   StructType public varStruct;


   function initialize() public {
       // Initialize variables outside the struct
       varString = "Outside Struct String";
       varUint = 54321;
       arrDynamicUint64.push(100);
       arrDynamicUint64.push(200);
       arrDynamicUint64.push(300);


       // Initialize the struct
       varStruct = StructType("Inside Struct String", 12345);
   }
}
