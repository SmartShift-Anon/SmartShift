// SPDX-License-Identifier: MIT
// OpenZeppelin Contracts (last updated v5.1.0) (token/ERC1155/IERC1155.sol)

pragma solidity ^0.8.20;



/**
 * @dev Required interface of an ERC-1155 compliant contract, as defined in the
 * https://eips.ethereum.org/EIPS/eip-1155[ERC].
 */
interface IERC1155{
    /**
     * @dev Emitted when `value` amount of tokens of type `id` are transferred from `from` to `to` by `operator`.
     */
    event TransferSingle(address indexed operator, address indexed from, address indexed to, uint256 id, uint256 value);

    /**
     * @dev Equivalent to multiple {TransferSingle} events, where `operator`, `from` and `to` are the same for all
     * transfers.
     */
    event TransferBatch(
        address indexed operator,
        address indexed from,
        address indexed to,
        uint256[] ids,
        uint256[] values
    );

    /**
     * @dev Emitted when `account` grants or revokes permission to `operator` to transfer their tokens, according to
     * `approved`.
     */
    event ApprovalForAll(address indexed account, address indexed operator, bool approved);

    /**
     * @dev Emitted when the URI for token type `id` changes to `value`, if it is a non-programmatic URI.
     *
     * If an {URI} event was emitted for `id`, the standard
     * https://eips.ethereum.org/EIPS/eip-1155#metadata-extensions[guarantees] that `value` will equal the value
     * returned by {IERC1155MetadataURI-uri}.
     */
    event URI(string value, uint256 indexed id);

    /**
     * @dev Returns the value of tokens of token type `id` owned by `account`.
     */
    function balanceOf(address account, uint256 id) external view returns (uint256);

    /**
     * @dev xref:ROOT:erc1155.adoc#batch-operations[Batched] version of {balanceOf}.
     *
     * Requirements:
     *
     * - `accounts` and `ids` must have the same length.
     */
    function balanceOfBatch(
        address[] calldata accounts,
        uint256[] calldata ids
    ) external view returns (uint256[] memory);

    /**
     * @dev Grants or revokes permission to `operator` to transfer the caller's tokens, according to `approved`,
     *
     * Emits an {ApprovalForAll} event.
     *
     * Requirements:
     *
     * - `operator` cannot be the zero address.
     */
    function setApprovalForAll(address operator, bool approved) external;

    /**
     * @dev Returns true if `operator` is approved to transfer ``account``'s tokens.
     *
     * See {setApprovalForAll}.
     */
    function isApprovedForAll(address account, address operator) external view returns (bool);

    /**
     * @dev Transfers a `value` amount of tokens of type `id` from `from` to `to`.
     *
     * WARNING: This function can potentially allow a reentrancy attack when transferring tokens
     * to an untrusted contract, when invoking {onERC1155Received} on the receiver.
     * Ensure to follow the checks-effects-interactions pattern and consider employing
     * reentrancy guards when interacting with untrusted contracts.
     *
     * Emits a {TransferSingle} event.
     *
     * Requirements:
     *
     * - `to` cannot be the zero address.
     * - If the caller is not `from`, it must have been approved to spend ``from``'s tokens via {setApprovalForAll}.
     * - `from` must have a balance of tokens of type `id` of at least `value` amount.
     * - If `to` refers to a smart contract, it must implement {IERC1155Receiver-onERC1155Received} and return the
     * acceptance magic value.
     */
    function safeTransferFrom(address from, address to, uint256 id, uint256 value, bytes calldata data) external;

    /**
     * @dev xref:ROOT:erc1155.adoc#batch-operations[Batched] version of {safeTransferFrom}.
     *
     * WARNING: This function can potentially allow a reentrancy attack when transferring tokens
     * to an untrusted contract, when invoking {onERC1155BatchReceived} on the receiver.
     * Ensure to follow the checks-effects-interactions pattern and consider employing
     * reentrancy guards when interacting with untrusted contracts.
     *
     * Emits either a {TransferSingle} or a {TransferBatch} event, depending on the length of the array arguments.
     *
     * Requirements:
     *
     * - `ids` and `values` must have the same length.
     * - If `to` refers to a smart contract, it must implement {IERC1155Receiver-onERC1155BatchReceived} and return the
     * acceptance magic value.
     */
    function safeBatchTransferFrom(
        address from,
        address to,
        uint256[] calldata ids,
        uint256[] calldata values,
        bytes calldata data
    ) external;
}

contract BasicERC1155 is IERC1155 {
    // Mapping from token ID to account balances
    mapping(uint256 => mapping(address => uint256)) private _balances;

    // Mapping from account to operator approvals
    mapping(address => mapping(address => bool)) private _operatorApprovals;

    function balanceOf(address account, uint256 id) external view returns (uint256) {
        require(account != address(0), "ERC1155: address zero query");
        return _balances[id][account];
    }

    function balanceOfBatch(
        address[] calldata accounts,
        uint256[] calldata ids
    ) external view returns (uint256[] memory) {
        require(accounts.length == ids.length, "ERC1155: accounts and ids length mismatch");

        uint256[] memory batchBalances = new uint256[](accounts.length);

        for (uint256 i = 0; i < accounts.length; i++) {
            batchBalances[i] = _balances[ids[i]][accounts[i]];
        }

        return batchBalances;
    }

    function setApprovalForAll(address operator, bool approved) external {
        require(msg.sender != operator, "ERC1155: setting approval status for self");
        _operatorApprovals[msg.sender][operator] = approved;
        emit ApprovalForAll(msg.sender, operator, approved);
    }

    function isApprovedForAll(address account, address operator) external view returns (bool) {
        return _operatorApprovals[account][operator];
    }

    function safeTransferFrom(
        address from,
        address to,
        uint256 id,
        uint256 value,
        bytes calldata data
    ) external {
        require(
            from == msg.sender || _operatorApprovals[from][msg.sender],
            "ERC1155: caller is not token owner or approved"
        );
        _safeTransferFrom(from, to, id, value, data);
    }

    function safeBatchTransferFrom(
        address from,
        address to,
        uint256[] calldata ids,
        uint256[] calldata values,
        bytes calldata data
    ) external {
        require(
            from == msg.sender || _operatorApprovals[from][msg.sender],
            "ERC1155: caller is not token owner or approved"
        );
        _safeBatchTransferFrom(from, to, ids, values, data);
    }

    // Internal functions
    function _safeTransferFrom(
        address from,
        address to,
        uint256 id,
        uint256 value,
        bytes memory data
    ) internal {
        require(to != address(0), "ERC1155: transfer to zero address");

        address operator = msg.sender;
        uint256[] memory ids = _asSingletonArray(id);
        uint256[] memory values = _asSingletonArray(value);

        _beforeTokenTransfer(operator, from, to, ids, values, data);

        uint256 fromBalance = _balances[id][from];
        require(fromBalance >= value, "ERC1155: insufficient balance for transfer");
        unchecked {
            _balances[id][from] = fromBalance - value;
        }
        _balances[id][to] += value;

        emit TransferSingle(operator, from, to, id, value);

        _afterTokenTransfer(operator, from, to, ids, values, data);

        _doSafeTransferAcceptanceCheck(operator, from, to, id, value, data);
    }

    function _safeBatchTransferFrom(
        address from,
        address to,
        uint256[] memory ids,
        uint256[] memory values,
        bytes memory data
    ) internal {
        require(ids.length == values.length, "ERC1155: ids and values length mismatch");
        require(to != address(0), "ERC1155: transfer to zero address");

        address operator = msg.sender;

        _beforeTokenTransfer(operator, from, to, ids, values, data);

        for (uint256 i = 0; i < ids.length; i++) {
            uint256 id = ids[i];
            uint256 value = values[i];

            uint256 fromBalance = _balances[id][from];
            require(fromBalance >= value, "ERC1155: insufficient balance for transfer");
            unchecked {
                _balances[id][from] = fromBalance - value;
            }
            _balances[id][to] += value;
        }

        emit TransferBatch(operator, from, to, ids, values);

        _afterTokenTransfer(operator, from, to, ids, values, data);

        _doSafeBatchTransferAcceptanceCheck(operator, from, to, ids, values, data);
    }

    function _mint(
        address to,
        uint256 id,
        uint256 value,
        bytes memory data
    ) internal {
        require(to != address(0), "ERC1155: mint to zero address");

        address operator = msg.sender;
        uint256[] memory ids = _asSingletonArray(id);
        uint256[] memory values = _asSingletonArray(value);

        _beforeTokenTransfer(operator, address(0), to, ids, values, data);

        _balances[id][to] += value;
        emit TransferSingle(operator, address(0), to, id, value);

        _afterTokenTransfer(operator, address(0), to, ids, values, data);

        _doSafeTransferAcceptanceCheck(operator, address(0), to, id, value, data);
    }

    function _mintBatch(
        address to,
        uint256[] memory ids,
        uint256[] memory values,
        bytes memory data
    ) internal {
        require(to != address(0), "ERC1155: mint to zero address");
        require(ids.length == values.length, "ERC1155: ids and values length mismatch");

        address operator = msg.sender;

        _beforeTokenTransfer(operator, address(0), to, ids, values, data);

        for (uint256 i = 0; i < ids.length; i++) {
            _balances[ids[i]][to] += values[i];
        }

        emit TransferBatch(operator, address(0), to, ids, values);

        _afterTokenTransfer(operator, address(0), to, ids, values, data);

        _doSafeBatchTransferAcceptanceCheck(operator, address(0), to, ids, values, data);
    }

    function _burn(
        address from,
        uint256 id,
        uint256 value
    ) internal {
        require(from != address(0), "ERC1155: burn from zero address");

        address operator = msg.sender;
        uint256[] memory ids = _asSingletonArray(id);
        uint256[] memory values = _asSingletonArray(value);

        _beforeTokenTransfer(operator, from, address(0), ids, values, "");

        uint256 fromBalance = _balances[id][from];
        require(fromBalance >= value, "ERC1155: burn amount exceeds balance");
        unchecked {
            _balances[id][from] = fromBalance - value;
        }

        emit TransferSingle(operator, from, address(0), id, value);

        _afterTokenTransfer(operator, from, address(0), ids, values, "");
    }

    function _asSingletonArray(uint256 element) private pure returns (uint256[] memory) {
        uint256[] memory array = new uint256[](1);
        array[0] = element;
        return array;
    }

    function _beforeTokenTransfer(
        address operator,
        address from,
        address to,
        uint256[] memory ids,
        uint256[] memory values,
        bytes memory data
    ) internal virtual {}

    function _afterTokenTransfer(
        address operator,
        address from,
        address to,
        uint256[] memory ids,
        uint256[] memory values,
        bytes memory data
    ) internal virtual {}

    function _doSafeTransferAcceptanceCheck(
        address operator,
        address from,
        address to,
        uint256 id,
        uint256 value,
        bytes memory data
    ) private {
        if (to.code.length > 0) {
            try IERC1155Receiver(to).onERC1155Received(operator, from, id, value, data) returns (bytes4 response) {
                if (response != IERC1155Receiver.onERC1155Received.selector) {
                    revert("ERC1155: ERC1155Receiver rejected tokens");
                }
            } catch Error(string memory reason) {
                revert(reason);
            } catch {
                revert("ERC1155: transfer to non-ERC1155Receiver implementer");
            }
        }
    }

    function _doSafeBatchTransferAcceptanceCheck(
        address operator,
        address from,
        address to,
        uint256[] memory ids,
        uint256[] memory values,
        bytes memory data
    ) private {
        if (to.code.length > 0) {
            try IERC1155Receiver(to).onERC1155BatchReceived(operator, from, ids, values, data) returns (bytes4 response) {
                if (response != IERC1155Receiver.onERC1155BatchReceived.selector) {
                    revert("ERC1155: ERC1155Receiver rejected tokens");
                }
            } catch Error(string memory reason) {
                revert(reason);
            } catch {
                revert("ERC1155: transfer to non-ERC1155Receiver implementer");
            }
        }
    }
}

interface IERC1155Receiver {
    function onERC1155Received(
        address operator,
        address from,
        uint256 id,
        uint256 value,
        bytes calldata data
    ) external returns (bytes4);

    function onERC1155BatchReceived(
        address operator,
        address from,
        uint256[] calldata ids,
        uint256[] calldata values,
        bytes calldata data
    ) external returns (bytes4);
}