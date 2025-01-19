// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;
//90320 gas
contract Contract3 {
   string[] public arrString;
   uint256 public varUint;


   function initialize() public {
       arrString.push("Item 1");
       arrString.push("Item 2");
       varUint = 100;
   }
}
