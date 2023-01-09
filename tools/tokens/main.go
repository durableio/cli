package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/durableio/cli/pkg/tokens"
	"github.com/google/uuid"
)

func main() {

	tm, err := tokens.Bootstrap()
	if err != nil {
		log.Fatal(err)
	}

	workflowId := fmt.Sprintf("wf_%s", strings.ReplaceAll(uuid.New().String(), "-", ""))

	token ,err := tm.CreateWorkflowToken(workflowId)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(token)
}
