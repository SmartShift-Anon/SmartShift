package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// ANSI escape codes for text colors
	black   = "\033[0;30m"
	red     = "\033[0;31m"
	green   = "\033[0;32m"
	yellow  = "\033[93m"
	orange  = "\033[38;5;208m"
	blue    = "\033[0;34m"
	magenta = "\033[0;35m"
	cyan    = "\033[0;36m"
	white   = "\033[0;37m"
	reset   = "\033[0m"
)

type StorageSlot struct {
	Key   common.Hash
	Value common.Hash
}

type DataElementInfo struct {
	Label  string      `json:"label"`
	Type   string      `json:"type"`
	Slot   common.Hash `json:"Slot"`
	Offset uint64      `json:"Offset"`
}

type Member struct {
	Offset uint64      `json:"Offset"`
	Slot   common.Hash `json:"Slot"`
	Type   string      `json:"type"`
}

type DataType struct {
	Type          string   `json:"type"`
	Base          string   `json:"base"`
	Encoding      string   `json:"encoding"`
	NumberOfBytes uint64   `json:"NumberOfBytes"`
	Members       []Member `json:"members"`
}

type ShardGenerator struct {
	commitedStorage  map[common.Hash]common.Hash
	extractedStorage map[common.Hash]common.Hash
	extractionInfos  []DataElementInfo
	dataTypes        map[string]DataType
}

func (s *ShardGenerator) Init(currentState map[common.Hash]common.Hash, extractionMessages []DataElementInfo, dataTypes []DataType) {

	s.commitedStorage = currentState
	s.extractionInfos = extractionMessages

	for _, dataType := range dataTypes {

		s.dataTypes[dataType.Type] = dataType
	}
}

func (s *ShardGenerator) GetCommitedState(key common.Hash) common.Hash {

	if _, ok := s.commitedStorage[key]; !ok {

		return common.Hash{}
	}

	return s.commitedStorage[key]

}

func (s *ShardGenerator) GetExtractedState(key common.Hash) common.Hash {

	if _, ok := s.extractedStorage[key]; !ok {

		return common.Hash{}
	}

	return s.extractedStorage[key]

}

func (s *ShardGenerator) SetExtractedState(key, val common.Hash) {

	s.extractedStorage[key] = val
}

func (s *ShardGenerator) IsStruct(dataType string) (bool, error) {

	if dataType, found := s.dataTypes[dataType]; found {

		if len(dataType.Members) == 0 {

			return false, nil

		} else {

			return true, nil
		}

	} else {

		return false, errors.New("Type not found")
	}
}

func (s *ShardGenerator) IsNested(dataType string) (bool, error) {

	if dataType, found := s.dataTypes[dataType]; found {

		if dataType.Base == "" {

			if len(dataType.Members) == 0 {

				return false, nil
			} else {

				return true, nil
			}

		} else {

			return true, nil
		}

	} else {

		return false, errors.New("Type not found")
	}
}

func (s *ShardGenerator) IsFlat(dataType string) (bool, error) {

	if dataType, found := s.dataTypes[dataType]; found {

		if dataType.Base == "" {

			return true, nil

		} else {

			return false, nil
		}

	} else {

		return false, errors.New("Type not found")
	}
}

func (s *ShardGenerator) IsEncodingInplace(dataType string) (bool, error) {

	if data, found := s.dataTypes[dataType]; found {

		if data.Encoding == "inplace" {

			return true, nil

		} else {

			return false, nil
		}

	} else {

		return false, errors.New("Type not found")
	}
}

func (s *ShardGenerator) IsEncodingDynamicArray(dataType string) (bool, error) {

	if data, found := s.dataTypes[dataType]; found {

		if data.Encoding == "dynamic_array" {

			return true, nil

		} else {

			return false, nil
		}

	} else {

		return false, errors.New("Type not found")
	}
}

func (s *ShardGenerator) IsEncodingBytes(dataType string) (bool, error) {

	if data, found := s.dataTypes[dataType]; found {

		if data.Encoding == "bytes" {

			return true, nil

		} else {

			return false, nil
		}

	} else {

		return false, errors.New("Type not found")
	}
}

func (s *ShardGenerator) GetNumberOfBytes(typeName string) (uint64, error) {

	if dataType, found := s.dataTypes[typeName]; found {

		return dataType.NumberOfBytes, nil

	} else {

		return 0, errors.New("Type not found")
	}
}

func WriteMapToJsonFile(data map[string]map[common.Hash]common.Hash, filename string) error {
	// Marshal the data into JSON
	jsonBytes, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	// Write the JSON data to a file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(jsonBytes)
	if err != nil {
		return err
	}

	log.Println("JSON file has been created successfully.")
	return nil
}

func (s *ShardGenerator) GenerateShards() (map[common.Hash]common.Hash, map[string]map[common.Hash]common.Hash, error) {

	mergedShardedState := make(map[common.Hash]common.Hash)
	shards := make(map[string]map[common.Hash]common.Hash)

	for _, reorgMessage := range s.extractionInfos {

		if isInplace, err := s.IsEncodingInplace(reorgMessage.Type); err != nil {

			return nil, nil, err

		} else if isInplace {
			err := s.ReorganizeInplace(reorgMessage)

			if err != nil {

				return nil, nil, err
			}

			shards[reorgMessage.Label] = s.extractedStorage
			s.extractedStorage = make(map[common.Hash]common.Hash)

		} else if isDynamicArray, err := s.IsEncodingDynamicArray(reorgMessage.Type); err != nil {

			return nil, nil, err

		} else if isDynamicArray {

			err := s.ReorganizeDynamicArray(reorgMessage)

			if err != nil {

				return nil, nil, err
			}

			shards[reorgMessage.Label] = s.extractedStorage
			s.extractedStorage = make(map[common.Hash]common.Hash)

		} else if isBytes, err := s.IsEncodingBytes(reorgMessage.Type); err != nil {

			return nil, nil, err

		} else if isBytes {

			err := s.ReorganizeBytes(reorgMessage)

			if err != nil {

				return nil, nil, err
			}

			shards[reorgMessage.Label] = s.extractedStorage
			s.extractedStorage = make(map[common.Hash]common.Hash)

		} else {

			return nil, nil, errors.New("Not implemented yet")
		}
	}

	for _, shard := range shards {

		for key, val := range shard {

			mergedShardedState[key] = val
		}
	}
	return mergedShardedState, shards, nil

}

func (s *ShardGenerator) ExtractUntilInplace(typeName string) (string, string, bool, error) {

	curType := typeName

	for {

		if dataType, found := s.dataTypes[curType]; found {

			if dataType.Base == "" {

				return dataType.Type, dataType.Encoding, false, nil

			} else {

				if dataType.Encoding != "inplace" {

					return dataType.Type, dataType.Encoding, true, nil

				} else {

					curType = dataType.Base
				}
			}

		} else {

			return "", "", false, errors.New("Type not found")
		}
	}
}

func (s *ShardGenerator) ContainsStruct(typeName string) (bool, string, error) {

	curType := typeName

	for {

		if dataType, found := s.dataTypes[curType]; found {

			if len(dataType.Members) != 0 {

				return true, curType, nil
			} else {

				if dataType.Base == "" {

					return false, "", nil
				} else {
					curType = dataType.Base
				}
			}
		} else {
			return false, "", errors.New("Type not found")
		}
	}
}

func (s *ShardGenerator) ReorganizeInplace(extractionMessage DataElementInfo) error {

	numberOfBytes, err := s.GetNumberOfBytes(extractionMessage.Type)

	if err != nil {

		return err
	}

	slotNumber := extractionMessage.Slot.Big()

	typeName, encoding, found, err := s.ExtractUntilInplace(extractionMessage.Type)

	if err != nil {

		return err
	}

	structFound, structTypeName, err := s.ContainsStruct(extractionMessage.Type)

	if err != nil {

		return err
	}

	if found {

		if encoding == "dynamic_array" {

			for i := 0; i < int(numberOfBytes/32); i++ {

				curSlotNumber := new(big.Int).Add(new(big.Int).SetInt64(int64(i)), slotNumber)

				err := s.ReorganizeDynamicArray(DataElementInfo{
					Type:   typeName,
					Slot:   common.BytesToHash(curSlotNumber.Bytes()),
					Offset: 0,
				})

				if err != nil {

					return err
				}
			}

		} else if encoding == "bytes" {

			for i := 0; i < int(numberOfBytes/32); i++ {

				curSlotNumber := new(big.Int).Add(new(big.Int).SetInt64(int64(i)), slotNumber)

				err := s.ReorganizeBytes(DataElementInfo{
					Type:   typeName,
					Slot:   common.BytesToHash(curSlotNumber.Bytes()),
					Offset: 0,
				})

				if err != nil {

					return err
				}
			}

		} else {

			return errors.New("Not implemented yet")
		}

		return nil

	} else if structFound {

		structSize, err := s.GetNumberOfBytes(structTypeName)
		if err != nil {
			return err
		}
		structDataType := s.dataTypes[structTypeName]

		curSlot := extractionMessage.Slot.Big()

		for i := 0; i < int(numberOfBytes)/int(structSize); i++ {

			for _, member := range structDataType.Members {

				memberDataType, exists := s.dataTypes[member.Type]
				if !exists {

					return errors.New("Struct Member Not Found")
				}

				if memberDataType.Encoding == "inplace" {

					err := s.ReorganizeInplace(DataElementInfo{
						Slot:   common.BigToHash(new(big.Int).Add(curSlot, member.Slot.Big())),
						Offset: member.Offset,
						Type:   memberDataType.Type,
					})

					if err != nil {

						return err
					}

				} else if memberDataType.Encoding == "dynamic_array" {

					err := s.ReorganizeDynamicArray(DataElementInfo{
						Slot:   common.BigToHash(new(big.Int).Add(curSlot, member.Slot.Big())),
						Offset: member.Offset,
						Type:   memberDataType.Type,
					})

					if err != nil {

						return err
					}

				} else if memberDataType.Encoding == "bytes" {

					err := s.ReorganizeBytes(DataElementInfo{
						Slot:   common.BigToHash(new(big.Int).Add(curSlot, member.Slot.Big())),
						Offset: member.Offset,
						Type:   memberDataType.Type,
					})

					if err != nil {

						return err
					}
				} else {

					return errors.New("Unknown Encoding")
				}
			}

			curSlot = new(big.Int).Add(curSlot, new(big.Int).SetUint64(structSize/32))
		}

		return nil

	} else {

		var offset uint64

		for offset = extractionMessage.Offset; offset < numberOfBytes+extractionMessage.Offset; offset = offset + 1 {

			curSlotNumber := new(big.Int).Add(new(big.Int).SetUint64(offset/32), slotNumber)

			slot := s.GetCommitedState(common.BytesToHash(curSlotNumber.Bytes()))
			shardSlot := s.GetExtractedState(common.BytesToHash(curSlotNumber.Bytes()))

			shardSlot[31-(offset%32)] = slot[31-(offset%32)]

			s.SetExtractedState(common.BytesToHash(curSlotNumber.Bytes()), shardSlot)

		}
		return nil
	}

}

func (s *ShardGenerator) ReorganizeDynamicArray(extractionMessage DataElementInfo) error {

	numberOfBytes, err := s.GetNumberOfBytes(extractionMessage.Type)

	if err != nil {

		return err
	}

	slot := s.GetCommitedState(extractionMessage.Slot)
	shardSlot := s.GetExtractedState(extractionMessage.Slot)

	for i := 0; i < int(numberOfBytes); i++ {

		shardSlot[i] = slot[i]
	}

	s.SetExtractedState(extractionMessage.Slot, shardSlot)

	dataSlot := common.BytesToHash(crypto.Keccak256(extractionMessage.Slot[:]))

	dataType := s.dataTypes[extractionMessage.Type]

	numberOfElements := slot.Big()

	if numberOfElements.Cmp(big.NewInt(0)) == 0 {

		return nil
	}

	if isInplace, err := s.IsEncodingInplace(dataType.Base); err != nil {

		return err

	} else if isInplace {

		if isNested, err := s.IsNested(dataType.Base); err != nil {

			return err

		} else if isNested {

			sizeOfElement, err := s.GetNumberOfBytes(dataType.Base)

			if err != nil {

				return err
			}

			numberOfSlotsPerElement := new(big.Int).SetUint64(sizeOfElement / 32)

			for i := big.NewInt(0); i.Cmp(numberOfElements) < 0; i.Add(i, big.NewInt(1)) {

				err := s.ReorganizeInplace(DataElementInfo{
					Slot:   common.BigToHash(new(big.Int).Add(dataSlot.Big(), new(big.Int).Mul(numberOfSlotsPerElement, i))),
					Offset: 0,
					Type:   dataType.Base,
				})

				if err != nil {

					return err
				}
			}

		} else if isFlat, err := s.IsFlat(dataType.Base); err != nil {

			return err

		} else if isFlat {

			sizeOfElement, err := s.GetNumberOfBytes(dataType.Base)

			if err != nil {

				return err
			}

			numberOfElementsPerSlot := new(big.Int).SetUint64(32 / sizeOfElement)

			numberOfSlots := big.NewInt(0)
			remainder := big.NewInt(0)

			numberOfSlots.DivMod(numberOfElements, numberOfElementsPerSlot, remainder)

			if remainder.Cmp(big.NewInt(0)) > 0 {

				numberOfSlots.Add(numberOfSlots, big.NewInt(1))
			}

			for i := big.NewInt(0); i.Cmp(numberOfSlots) < 0; i.Add(i, big.NewInt(1)) {

				for j := uint64(0); j < 32/sizeOfElement; j++ {

					err := s.ReorganizeInplace(DataElementInfo{
						Slot:   common.BigToHash(new(big.Int).Add(dataSlot.Big(), i)),
						Offset: j * sizeOfElement,
						Type:   dataType.Base,
					})

					if err != nil {

						return err
					}
				}
			}

		} else {

			return errors.New("Not Implemented Yet....")
		}

	} else if isDynamicArray, err := s.IsEncodingDynamicArray(dataType.Base); err != nil {

		return err

	} else if isDynamicArray {

		for i := big.NewInt(0); i.Cmp(numberOfElements) < 0; i.Add(i, big.NewInt(1)) {

			err := s.ReorganizeInplace(DataElementInfo{
				Slot:   common.BigToHash(new(big.Int).Add(dataSlot.Big(), i)),
				Offset: 0,
				Type:   dataType.Base,
			})

			if err != nil {

				return err
			}
		}

	} else if isBytes, err := s.IsEncodingBytes(dataType.Base); err != nil {

		return err

	} else if isBytes {

		for i := big.NewInt(0); i.Cmp(numberOfElements) < 0; i.Add(i, big.NewInt(1)) {

			err := s.ReorganizeInplace(DataElementInfo{
				Slot:   common.BigToHash(new(big.Int).Add(dataSlot.Big(), i)),
				Offset: 0,
				Type:   dataType.Base,
			})

			if err != nil {

				return err
			}
		}

	} else {

		return errors.New("Not Implemented Yet....")
	}

	return nil

}

func (s *ShardGenerator) ReorganizeBytes(extractionMessage DataElementInfo) error {

	numberOfBytes, err := s.GetNumberOfBytes(extractionMessage.Type)

	if err != nil {

		return err
	}

	slot := s.GetCommitedState(extractionMessage.Slot)
	shardSlot := s.GetExtractedState(extractionMessage.Slot)

	for i := 0; i < int(numberOfBytes); i++ {

		shardSlot[i] = slot[i]
	}

	s.SetExtractedState(extractionMessage.Slot, shardSlot)

	dataSlot := common.BytesToHash(crypto.Keccak256(extractionMessage.Slot[:]))

	if (slot[31] & 1) != 0 {

		numberOfElements := new(big.Int).Div(new(big.Int).Sub(slot.Big(), big.NewInt(1)), big.NewInt(2))
		numberOfSlots := big.NewInt(0)
		remainder := big.NewInt(0)
		numberOfSlots.DivMod(numberOfElements, big.NewInt(32), remainder)

		if remainder.Cmp(big.NewInt(0)) > 0 {

			numberOfSlots.Add(numberOfSlots, big.NewInt(1))
		}

		for i := big.NewInt(0); i.Cmp(numberOfSlots) < 0; i.Add(i, big.NewInt(1)) {

			slotToBeCopiedFrom := common.BigToHash(new(big.Int).Add(dataSlot.Big(), i))
			slotToBeCopiedTo := slotToBeCopiedFrom

			curSlot := s.GetCommitedState(slotToBeCopiedFrom)
			curShardSlot := s.GetExtractedState(slotToBeCopiedTo)

			for j := 0; j < 32; j++ {

				curShardSlot[j] = curSlot[j]
			}

			s.SetExtractedState(slotToBeCopiedTo, curShardSlot)

		}

	}

	return nil

}

func NewShardGenerator() *ShardGenerator {
	return &ShardGenerator{
		commitedStorage:  make(map[common.Hash]common.Hash),
		extractedStorage: make(map[common.Hash]common.Hash),
		dataTypes:        make(map[string]DataType),
	}
}

func ReadStorageFromFile(filePath string) (*map[common.Hash]StorageSlot, error) {
	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println(red + err.Error() + reset)
		return nil, err
	}

	defer file.Close()

	byteVal, _ := ioutil.ReadAll(file)
	var storageSlots map[common.Hash]StorageSlot
	json.Unmarshal(byteVal, &storageSlots)
	for key, slot := range storageSlots {

		if slot.Value.Cmp(common.Hash{}) == 0 {
			delete(storageSlots, key)
		}
	}
	return &storageSlots, nil
}

func ReadDataElemFromFile(filePath string) ([]DataElementInfo, error) {

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println(red + err.Error() + reset)
		return nil, err
	}

	defer file.Close()

	byteVal, _ := ioutil.ReadAll(file)
	var reorgInfos []DataElementInfo
	json.Unmarshal(byteVal, &reorgInfos)
	return reorgInfos, nil
}

func ReadDataTypesFromFile(filePath string) ([]DataType, error) {

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println(red + err.Error() + reset)
		return nil, err
	}

	defer file.Close()

	byteVal, _ := ioutil.ReadAll(file)
	var dataTypes []DataType
	json.Unmarshal(byteVal, &dataTypes)
	return dataTypes, nil
}

func isEqual(storage map[common.Hash]common.Hash, shards map[common.Hash]common.Hash) error {

	for key, val := range storage {

		otherVal, found := shards[key]

		if !found {

			return errors.New("key not found")
		}

		if !bytes.Equal(val[:], otherVal[:]) {

			return errors.New("value mismatch")
		}
	}

	for key, val := range shards {

		otherVal, found := storage[key]

		if !found {

			return errors.New("key not found")
		}

		if !bytes.Equal(val[:], otherVal[:]) {

			return errors.New("value mismatch")
		}
	}

	return nil
}

func runTest(directoryPath string) (bool, error) {

	fmt.Println(cyan + "Current Directory: " + directoryPath + reset)

	storageSlots, err := ReadStorageFromFile(directoryPath + "/" + "old_storage.json")

	if err != nil {
		fmt.Println(red + err.Error() + reset)
		return false, err
	}

	reorgInfos, err := ReadDataElemFromFile(directoryPath + "/" + "storage_reorg_info.json")

	if err != nil {

		fmt.Println(red + err.Error() + reset)
		return false, err
	}

	dataTypes, err := ReadDataTypesFromFile(directoryPath + "/" + "data_types.json")

	if err != nil {

		fmt.Println(red + err.Error() + reset)
		return false, err
	}

	currentStateAsMap := make(map[common.Hash]common.Hash)
	for _, slot := range *storageSlots {

		currentStateAsMap[slot.Key] = slot.Value
	}
	reorganizer := NewShardGenerator()

	reorganizer.Init(currentStateAsMap, reorgInfos, dataTypes)

	mergedShards, shards, err := reorganizer.GenerateShards()

	WriteMapToJsonFile(shards, directoryPath+"/shards.json")

	if err != nil {
		return false, err
	}

	err = isEqual(currentStateAsMap, mergedShards)

	if err != nil {

		return false, err
	}

	fmt.Println(green + "Test passed: " + directoryPath + "🎉🎉🎉" + reset)
	return true, nil
}

func main() {

	runTest("Tests/test6")

}
