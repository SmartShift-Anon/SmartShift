import subprocess
import json
import sys
import os
import re

def int_to_256bit_hex_string(num):
    # Convert the integer to a hex string
    hex_string = hex(num)[2:]

    # Ensure the hex string is 256 bits long
    padded_hex_string = hex_string.zfill(64)

    # Add the "0x" prefix
    final_hex_string = "0x" + padded_hex_string

    return final_hex_string

def get_storage_layout(file_name):
    # Command to run (solc --stage-layout sample.sol)
    command = ["solc", "--storage-layout"]
    command.append(file_name)

    # Run the command
    result = subprocess.run(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, universal_newlines=True)

    # Check if the command was successful
    if result.returncode == 0:
        output = result.stdout

        # Find the index of "Contract Storage Layout:"
        layout_start_index = output.find("Contract Storage Layout:")

        # Extract the data after "Contract Storage Layout:"
        if layout_start_index != -1:
            layout_data_str = output[layout_start_index + len("Contract Storage Layout:"):].strip()

            # Convert the string to a dictionary using json.loads
            try:
                layout_data_dict = json.loads(layout_data_str)
                return layout_data_dict
            except json.JSONDecodeError as e:
                raise Exception("Error decoding JSON")
        else:
            raise Exception("Contract Storage Layout not found in the output.")
    else:
        raise Exception(result.stderr)




def get_objects(storage_json):
    storage = storage_json["storage"] #storage in the contract
    types = storage_json["types"] #data types in the contract
    
    storage_objects = []

    for storage_object in storage:
        storage_objects.append({
            "label":storage_object["label"],
            "type":storage_object["type"],
            "slot":int_to_256bit_hex_string(int(storage_object["slot"])),
            "offset":storage_object["offset"],
            })
    return storage_objects

"""
def get_types(old_json, common_objects):
    old_types = old_json["types"]
    inserted_types = []
    nested_types = []
    flat_types = []
    for common_object in common_objects:
        current_type = common_object["type"]
        while old_types.get(current_type,None) is not None and current_type not in inserted_types:
            old_types[current_type]["type"] = current_type
            old_types[current_type]["numberOfBytes"] = int(old_types[current_type]["numberOfBytes"])
            inserted_types.append(current_type)
            if "base" in old_types[current_type]:
                nested_types.append(old_types[current_type])
                current_type = old_types[current_type]["base"]
            else:
                flat_types.append(old_types[current_type])
                break
        if old_types.get(current_type,None) is None:
            raise Exception("Type not found....")
    return nested_types,flat_types
"""
def process_type(types, current_type, inserted_types, data_types):
    
    if current_type in inserted_types:
        return
    
    types[current_type]["type"] = current_type 
    types[current_type]["numberOfBytes"] = int(types[current_type]["numberOfBytes"]) 
    

    inserted_types.append(current_type)
    #if there is a base type process it too
    if "base" in types[current_type]:
        process_type(types,types[current_type]["base"],inserted_types,data_types)
    else:
        types[current_type]["base"] = None

    #if the data type is a struct then process the members
    if "members" in types[current_type]:
        members = {}
        
        
        for member in types[current_type]["members"]:
            members[member["label"]] = member

        
        
        for member in types[current_type]["members"]:
            process_type(types,member["type"],inserted_types,data_types)
    else:
        types[current_type]["members"] = None


    data_types.append(types[current_type])    

#find the data types of the storage objects that require reorganization
def get_types(types, storage_objects):
    inserted_types = []
    data_types = []
    
    for storage_object in storage_objects:
        current_type = storage_object["type"]
        process_type(types,current_type,inserted_types,data_types)
        
    
    for type in data_types:
        if type["members"] is not None:
            for member in type["members"]:
                
                member["slot"] = int_to_256bit_hex_string(int(member["slot"]))
                member.pop("astId")
                member.pop("contract")
                member.pop("label")
                
    return data_types

def writeJSON(file_name,data):
    with open(file_name, 'w') as json_file:
        json.dump(data, json_file, indent=2)


def get_directories_in_path(directory_path):
    # Get all files and directories in the specified path
    entries = os.listdir(directory_path)

    # Filter out directories
    directories = [entry for entry in entries if os.path.isdir(os.path.join(directory_path, entry))]

    return directories

#modify struct type name
def modify_struct_types(text):
    pattern = r't_struct\((.*?)\)[a-zA-Z0-9]+_storage'
    result = re.sub(pattern, r't_struct(\1)_storage', text)
    return result

def clean_types(storage_layout):
    storage = storage_layout["storage"]
    types = storage_layout["types"]

    for item in storage:
        modified_type_def = modify_struct_types(item["type"])
        item["type"] = modified_type_def

    all_keys = list(types.keys())
    for key in all_keys:
        new_key = modify_struct_types(key)
        if "base" in types[key]:
            base = types[key]["base"]
            new_base = modify_struct_types(base)
            if new_base != base:
                types[key]["base"] = new_base
        if new_key != key:
            types[new_key] = types.pop(key)

    storage_layout["storage"] = storage
    storage_layout["types"] = types



if __name__ == "__main__":
    
    target_directory = "Tests"
    directories_list = get_directories_in_path(target_directory)

    # Print the result
    for directory in directories_list:
        current_directory = target_directory+"/"+directory
        src_file = current_directory+"/"+"Sample.sol"
        storage_layout = get_storage_layout(src_file)
        clean_types(storage_layout)
        
        result = get_objects(storage_layout)
        
        data_types = get_types(storage_layout["types"],result)
        writeJSON(current_directory+"/"+"storage_reorg_info.json",result)
        writeJSON(current_directory+"/"+"data_types.json",data_types)

        
        """
        #nested,flat = get_types(old_storage_layout,result)
        data_types = get_types(old_storage_layout["types"],new_storage_layout["types"],result)
        #print(json.dumps(data_types,indent=2))
        
        writeJSON(current_directory+"/"+"storage_reorg_info.json",result)
        #writeJSON(current_directory+"/"+"nested_types.json",nested)
        #writeJSON(current_directory+"/"+"flat_types.json",flat)
        writeJSON(current_directory+"/"+"data_types.json",data_types)
        """
        