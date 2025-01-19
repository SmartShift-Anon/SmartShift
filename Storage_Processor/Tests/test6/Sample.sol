// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;
//291444
contract StructuredContract {
    struct ComplexStruct {
        string varString;
        uint256 varUint;
        bool varBool;
        uint256[] dynamicUintArray;
        address[5] fixedAddressArray;
        bytes varBytes;
    }

    ComplexStruct public complexVar;

    function initialize() public {
        complexVar.varString = "Hello, Solidity!";
        complexVar.varUint = 123456;
        complexVar.varBool = true;

        // Properly initialize a dynamic array with an initial size
        complexVar.dynamicUintArray = new uint256[](3) ;
        complexVar.dynamicUintArray[0] = 100;
        complexVar.dynamicUintArray[1] = 200;
        complexVar.dynamicUintArray[2] = 300;

        complexVar.fixedAddressArray = [address(0x1), address(0x2), address(0x3), address(0x4), address(0x5)];

        complexVar.varBytes = "Solidity bytes";

    }
}
