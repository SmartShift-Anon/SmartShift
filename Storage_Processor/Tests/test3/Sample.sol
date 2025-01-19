// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;
//136261 gas
contract Contract5 {
   uint64 public varUint64_1;
   uint256 public varUint256;
   uint64 public varUint64_2;
   uint64[4] public arrFixedUint64;
   uint64[] public arrDynamicUint64;


   function initialize() public {
       varUint64_1 = 1234567890;
       varUint256 = 9876543210;
       varUint64_2 = 1122334455;
       arrFixedUint64 = [10, 20, 30, 40];
       arrDynamicUint64.push(50);
       arrDynamicUint64.push(60);
       arrDynamicUint64.push(70);
   }
}
