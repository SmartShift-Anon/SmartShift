import subprocess
import json
import sys
import os
import re
import solcast
import solcx
from collections import deque
import requests




def create_solc_input_json(contract_file):
    """
    Reads a Solidity file and creates a JSON input for the Solidity compiler.

    Args:
    - contract_file (str): The file path to the Solidity contract.

    Returns:
    - str: A JSON string configured for use with the Solidity compiler, or
      an error message if the file could not be found.
    """
    # Read the contents of the Solidity file
    try:
        with open(contract_file, 'r') as file:
            content = file.read()
    except FileNotFoundError:
        return f"Error: The file '{contract_file}' does not exist."

    # Define the input dictionary for the compiler
    input_dict = {
        "language": "Solidity",
        "sources": {
            contract_file: {
                "content": content
            }
        },
        "settings": {
            "outputSelection": {
                "*": {
          "*": [
            "abi",
            "evm.bytecode",
            "evm.deployedBytecode",
            "evm.methodIdentifiers",
            "metadata"
          ],
          "": [
            "ast"
          ]
        }

            }
        }
    }

    # Convert the Python dictionary to a JSON string
    input_json = json.dumps(input_dict, indent=4)
    
    return input_json





def construct_function_dependency_graph(contract_def):
    dependency_graph = {}

    function_defs = contract_def.children(
        include_children=False,
        filters={'nodeType': "FunctionDefinition", 'kind':"function"}
    )
    
    
    for function_def in function_defs:
        dependency_graph[function_def.id] = set()
        function_calls = function_def.children(
            include_children=False,
            filters={'nodeType': "FunctionCall"}
            )
        if len(function_calls) > 0:
            for function_call in function_calls:
                if function_call.expression.nodeType == "Identifier" and any(func_def.id == function_call.expression.referencedDeclaration for func_def in function_defs) and (function_def.id != function_call.expression.referencedDeclaration):
                    dependency_graph[function_def.id].add(function_call.expression.referencedDeclaration)
                    

    return dependency_graph


def construct_state_dependency_matrix(contract_def):
    state_dependency_matrix = {}
    state_vars = contract_def.children(
        include_children=False,
        filters={'nodeType': "VariableDeclaration",'stateVariable':True}
    )

    function_defs = contract_def.children(
        include_children=False,
        filters={'nodeType': "FunctionDefinition", 'kind':"function"}
    )

    for function_def in function_defs:
        state_dependency_matrix[function_def.id] = set()
        identifiers = function_def.children(
            include_children=False,
            filters={'nodeType': "Identifier"}
        )
        for identifier in identifiers:
            if any(state_var.id == identifier.referencedDeclaration for state_var in state_vars):
                state_dependency_matrix[function_def.id].add(identifier.referencedDeclaration)

    return state_dependency_matrix    
    

def detect_cycle(graph):
    visited = set()
    rec_stack = set()

    def dfs(node):
        if node in rec_stack:
            return True  # Cycle detected
        if node in visited:
            return False

        visited.add(node)
        rec_stack.add(node)
        
        # Recur for all vertices adjacent to this vertex
        for neighbor in graph[node]:
            if dfs(neighbor):
                return True

        rec_stack.remove(node)
        return False

    for node in graph:
        if node not in visited:
            if dfs(node):
                return True
    return False

def topological_sort(graph):
    in_degree = {node: 0 for node in graph}
    for node in graph:
        for neighbor in graph[node]:
            in_degree[neighbor] += 1
    
    queue = deque([node for node in in_degree if in_degree[node] == 0])
    topo_order = []

    while queue:
        vertex = queue.popleft()
        topo_order.append(vertex)
        
        for neighbor in graph[vertex]:
            in_degree[neighbor] -= 1
            if in_degree[neighbor] == 0:
                queue.append(neighbor)
    
    if len(topo_order) == len(graph):
        topo_order.reverse()
        return topo_order
    else:
        return None  # This means there was a cycle, which should not happen here


def modify_state_dependency_matrix(func_sorted_order, func_dependency_matrix, state_dependency_matrix):

    for func in func_sorted_order:
        for dpendent_func in func_dependency_matrix[func]:
            state_dependency_matrix[func] = state_dependency_matrix[func] | state_dependency_matrix[dpendent_func]

    return state_dependency_matrix



def calculate_function_activation_threshold(contract_def,state_dependency_matrix):

    var_decl = contract_def.children(
        include_children=False,
        filters={'nodeType': "VariableDeclaration",'stateVariable':True}
    )
    
    count_of_vars = 0
    count_of_funcs = 0
    
    for func,state_vars in state_dependency_matrix.items():
        count_of_funcs = count_of_funcs + 1
        count_of_vars = count_of_vars + len(state_vars)

    print(len(var_decl))        
    print(count_of_vars/count_of_funcs)    

def construct_readable_function_dependency_matrix(contract_def,func_dependency_matrix):
    readable_func_dependency_matrix = {}

    for func,dependencies in func_dependency_matrix.items():
        
        function_def = contract_def.children(
            include_children=False,
            filters={'nodeType': "FunctionDefinition", 'id':func}
        )

        readable_func_dependency_matrix[function_def[0].name] = set()

        for dependency in dependencies:

            dependency_def = contract_def.children(
                include_children=False,
                filters={'nodeType': "FunctionDefinition", 'id':dependency}
            )

            readable_func_dependency_matrix[function_def[0].name].add(dependency_def[0].name)

    
    return readable_func_dependency_matrix



def construct_readable_state_dependency_matrix(contract_def,state_dependency_matrix):
    readable_state_dependency_matrix = {}

    for func,dependencies in state_dependency_matrix.items():
        
        function_def = contract_def.children(
            include_children=False,
            filters={'nodeType': "FunctionDefinition", 'id':func}
        )

        readable_state_dependency_matrix[function_def[0].name] = set()

        for dependency in dependencies:

            dependency_def = contract_def.children(
                include_children=False,
                filters={'nodeType': "VariableDeclaration",'stateVariable':True,'id':dependency}
            )

            readable_state_dependency_matrix[function_def[0].name].add(dependency_def[0].name)

    
    return readable_state_dependency_matrix

contract_filename = 'Tokens/ERC-1155/ERC-1155.sol'
json_input = create_solc_input_json(contract_filename)
output_json = solcx.compile_standard(json.loads(json_input))
source_units = solcast.from_standard_output(output_json)
contract_defs = source_units[0].children(filters={'nodeType': "ContractDefinition",'contractKind':"contract"})

for contract_def in contract_defs:   
    func_dependency_matrix = construct_function_dependency_graph(contract_def)
    #readable_func_dependency_matrix = construct_readable_function_dependency_matrix(contract_def,func_dependency_matrix)
    #print(readable_func_dependency_matrix)
    state_dependency_matrix = construct_state_dependency_matrix(contract_def)
    if detect_cycle(func_dependency_matrix):
        raise ValueError("Cycle detected in the graph")
    
    sorted_order = topological_sort(func_dependency_matrix)
    state_dependency_matrix = modify_state_dependency_matrix(sorted_order,func_dependency_matrix,state_dependency_matrix)
    readable_state_dependency_matrix = construct_readable_state_dependency_matrix(contract_def,state_dependency_matrix)
    print(readable_state_dependency_matrix)
    
    





