import json
import sys
import os
import requests
import solcast
import solcx
from Crypto.Hash import keccak


def choose_contract(contracts):
    print("Available contracts:")
    contract_names = list(contracts.keys())
    for i, name in enumerate(contract_names):
        print(f"{i + 1}. {name}")

    # Get user input
    choice = int(input("Enter the number of the contract you want to choose: ")) - 1

    # Ensure the input is valid
    if choice < 0 or choice >= len(contract_names):
        print("Invalid choice, please run the program again.")
        return None

    selected_contract = contract_names[choice]
    return contracts[selected_contract]


def get_keccak_hash(input_string):
    k = keccak.new(digest_bits=256)
    k.update(input_string.encode())
    return k.hexdigest()

def extract_function_selectors(abi):
    selectors = {}
    for item in abi:
        if item['type'] == 'function':
            # Construct the function signature
            inputs = ','.join([input['type'] for input in item['inputs']])
            function_signature = f"{item['name']}({inputs})"
            # Compute the selector
            hash_digest = get_keccak_hash(function_signature)
            selector = hash_digest[:8]  # First 4 bytes (8 hex characters)
            selectors[selector] = item['name']
    return selectors


def get_slectors(contract_file):
    
    try:
        with open(contract_file, 'r') as file:
            content = file.read()
    except FileNotFoundError:
        return f"Error: The file '{contract_file}' does not exist."

    output = solcx.compile_source(content, output_values = ["abi"])
    selectors = {}
    for contract,abi in output.items():
        fun_selectors = extract_function_selectors(abi["abi"])
        selectors[contract.split(":")[1]] = fun_selectors

    return selectors




def get_latest_block_number():
    
    #api_key = "PUT_YOUR_KEY_HERE"
    url = f"https://mainnet.infura.io/v3/{api_key}"
    headers = {'Content-Type': 'application/json'}
    data = {
        "jsonrpc": "2.0",
        "method": "eth_blockNumber",
        "params": [],
        "id": 1
    }
    
    # Convert the data dictionary to JSON format
    json_data = json.dumps(data)

    # Making a POST request
    response = requests.post(url, headers=headers, data=json_data)

    # Check if the request was successful
    if response.status_code == 200:
        response_json = response.json()
        block_number_hex = response_json.get('result', None)
        if block_number_hex:
            return int(block_number_hex, 16)  # Convert hex to int
        return None
    else:
        print("Failed to retrieve data")
        print("Status Code:", response.status_code)
        print("Response Text:", response.text)
        return None








def fetch_selectors_from_transactions(address, latest_block_number):
    
    url = "https://api.etherscan.io/v2/api"

    #api_key = "PUT_YOUR_KEY_HERE"
    
    # Set the parameters for the GET request
    params = {
        "chainid": 1,
        "module": "account",
        "action": "txlist",
        "address": address,
        "startblock": latest_block_number-100,
        "endblock": latest_block_number,
        "page": 1,
        "offset": 100,
        "sort": "asc",
        "apikey": api_key
    }

    # Send the GET request
    response = requests.get(url, params=params)

    # Check if the request was successful
    if response.status_code == 200:
        # Parse the JSON response
        data = response.json()
        # Extract 'result' which contains the transactions
        transactions = data.get("result", [])
        
        # Collect all 'input' fields from transactions
        inputs = [tx["input"].split("x")[1][:8] for tx in transactions if "input" in tx]
        return inputs
    else:
        print("Failed to retrieve data: Status code", response.status_code)
        return []


address = "0x6982508145454Ce325dDbE47a25d4ec3d2311933"



contract_filename = 'sample.sol'
selectors = get_slectors(contract_filename)
selectors_in_contract = choose_contract(selectors)




latest_block_number = get_latest_block_number()
if latest_block_number is None:
    raise ValueError("Latest Block Number could not be fetched")
    
txs = fetch_selectors_from_transactions(address, latest_block_number)


prio_vec = {}
for selector,name in selectors_in_contract.items():
    prio_vec[name] = 0
for tx in txs:
    prio_vec[selectors_in_contract[tx]] = prio_vec[selectors_in_contract[tx]]+1

print(prio_vec)