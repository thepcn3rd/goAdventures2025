package main

import "fmt"

// Based the offline lookup on the k-Anonymity Model
type HashOfflineLookupStruct struct {
	SHA1HashPrefix []PrefixStruct `json:"sha1Prefix"`
	NTLMHashPrefix []PrefixStruct `json:"ntlmPrefix"`
}

type PrefixStruct struct {
	Prefix      string   `json:"prefix"`
	Suffix      []string `json:"suffix"`
	Created     string   `json:"created"`
	LastUpdated string   `json:"lastUpdated"`
}

func removeDuplicateSuffix(suffix []string) []string {

	seen := make(map[string]bool)
	result := []string{}

	for _, str := range suffix {
		if _, ok := seen[str]; !ok {
			seen[str] = true
			result = append(result, str)
		}
	}
	return result
}

func main() {
	var hashStruct HashOfflineLookupStruct
	var prefix PrefixStruct
	var prefix2 PrefixStruct
	var prefix3 PrefixStruct

	prefix.Prefix = "1CCDC"
	prefix.Suffix = append(prefix.Suffix, "B8E68B92E79CE344C25F3D87FC297D12346")
	prefix.Suffix = append(prefix.Suffix, "B8E68B92E79CE344C25F3D87FC297D12345")
	prefix.Suffix = append(prefix.Suffix, "B8E68B92E79CE344C25F3D87FC297D12344")
	prefix.Suffix = append(prefix.Suffix, "B8E68B92E79CE344C25F3D87FC297D12343")
	prefix.Suffix = append(prefix.Suffix, "B8E68B92E79CE344C25F3D87FC297D12342")
	prefix.Suffix = append(prefix.Suffix, "B8E68B92E79CE344C25F3D87FC297D12341")

	prefix2.Prefix = "1CCDC"
	prefix2.Suffix = append(prefix2.Suffix, "B8E68B92E79CE344C25F3D87FC297D12340")
	prefix2.Suffix = append(prefix2.Suffix, "B8E68B92E79CE344C25F3D87FC297D12347")
	prefix2.Suffix = append(prefix2.Suffix, "B8E68B92E79CE344C25F3D87FC297D12346")

	prefix3.Prefix = "1CCDD"
	prefix3.Suffix = append(prefix3.Suffix, "B8E68B92E79CE344C25F3D87FC297D12340")
	prefix3.Suffix = append(prefix3.Suffix, "B8E68B92E79CE344C25F3D87FC297D12347")
	prefix3.Suffix = append(prefix3.Suffix, "B8E68B92E79CE344C25F3D87FC297D12346")

	hashStruct.SHA1HashPrefix = append(hashStruct.SHA1HashPrefix, prefix)
	hashStruct.SHA1HashPrefix = append(hashStruct.SHA1HashPrefix, prefix3)
	prefixExists := false
	for i, h := range hashStruct.SHA1HashPrefix {
		if h.Prefix == prefix2.Prefix {
			fmt.Println("Prefix Exists")
			prefixExists = true
			hashStruct.SHA1HashPrefix[i].Suffix = append(hashStruct.SHA1HashPrefix[i].Suffix, prefix2.Suffix...)
			hashStruct.SHA1HashPrefix[i].Suffix = removeDuplicateSuffix(hashStruct.SHA1HashPrefix[i].Suffix)
		}
	}
	if !prefixExists {
		hashStruct.SHA1HashPrefix = append(hashStruct.SHA1HashPrefix, prefix2)
	}

	fmt.Println(hashStruct)
}
