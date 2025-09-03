package blocks

// 	When requiring dependencies on a YAML file the user will just pass the repo,
//  and it will be assumed that a agentic_support.yaml file is at the root of the project.
//  The function will fetch the agentic_support.yaml from the repository root.

func FetchBlock(repo string) (*BlockInfo, error) {
	blockyaml, err := getBlockFromRepo(repo)
	if err != nil {
		return nil, err
	}

	binaryPath, err := downloadAndStoreBinary(blockyaml)
	if err != nil {
		return nil, err
	}
	blockyaml.BinaryPath = binaryPath

	return blockyaml, nil
}
