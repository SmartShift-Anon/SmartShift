import json
import requests


def get_latest_block_number():
    
    api_key = "YOUR_API_KEY"
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
            return block_number_hex
        return None
    else:
        print("Failed to retrieve data")
        print("Status Code:", response.status_code)
        print("Response Text:", response.text)
        return None



def fetch_gas_limit(block_number):
    
    api_key = 'YOUR_API_KEY'

    # API endpoint
    url = f'https://mainnet.infura.io/v3/{api_key}'

    # JSON-RPC request payload
    payload = {
        "jsonrpc": "2.0",
        "method": "eth_getBlockByNumber",
        "params": [block_number, False],
        "id": 1
    }

    # Headers
    headers = {
        'Content-Type': 'application/json'
    }

    # POST request
    response = requests.post(url, json=payload, headers=headers)

    # Check if the request was successful
    if response.status_code == 200:
        data = response.json()
        gas_limit = data['result']['gasLimit']
        return int(gas_limit, 16)
    else:
        print("Failed to fetch data")
        return None



def extract_sorted_vars(prio_vec_file, dp_mat_file):
    
    with open(prio_vec_file, 'r') as file:
        prio_vec = json.load(file)

    
    with open(dp_mat_file, 'r') as file:
        dp_mat = json.load(file)

    
    sorted_funcs = sorted(prio_vec, key=lambda x: prio_vec[x], reverse=True)

    
    unique_vars = []

    
    for key in sorted_funcs:
        if key in dp_mat:
            for value in dp_mat[key]:
                if value not in unique_vars:
                    unique_vars.append(value)

    return unique_vars




def create_function_batches(data, unique_vars, dependency_matrix, batch_size):

    current_batch_size = 0
    current_batch = {}
    current_batch["activate"] = []
    funcs = list(dependency_matrix.keys())
    var_completed = {}
    for var in unique_vars:
        var_completed[var] = False
    batches = []
    for var in unique_vars:
        slots = data[var]
        for key,val in slots.items():
            current_batch[key] = val
            current_batch_size = current_batch_size+1
            if current_batch_size == batch_size:
                batches.append(current_batch)
                current_batch = {}
                current_batch["activate"] = []
                current_batch_size = 0

        var_completed[var] = True
        for func in funcs:
            dpendents = dependency_matrix[func]
            all_included = True
            for dependent in dpendents:
                if var_completed[dependent] == False:
                    all_included = False
                    break
            if all_included == True:
                current_batch["activate"].append(func)
                func.remove(func)









block_number = get_latest_block_number()
gas_limit = fetch_gas_limit(block_number)
batch_size = gas_limit/30000

