/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import "github.com/durableio/cli/cmd"

//--go:generate sqlboiler --wipe --output=gen/database --pkgname=database --add-global-variants mysql

func main() {
	cmd.Execute()
}
