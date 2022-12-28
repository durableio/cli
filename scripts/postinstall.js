#!/usr/bin/env node

// Ref 1: https://github.com/sanathkr/go-npm
// Ref 2: https://blog.xendit.engineer/how-we-repurposed-npm-to-publish-and-distribute-our-go-binaries-for-internal-cli-23981b80911b
"use strict";

import binLinks from "bin-links";
import fs from "fs";
import fetch from "node-fetch";
import path from "path";
import tar from "tar";
import zlib from "zlib";

// Mapping from Node's `process.arch` to Golang's `$GOARCH`
const ARCH_MAPPING = {
    x64: "amd64",
    arm64: "arm64",
};

// Mapping between Node's `process.platform` to Golang's
const PLATFORM_MAPPING = {
    darwin: "darwin",
    linux: "linux",
    win32: "windows",
};

// TODO: import pkg from "../package.json" assert { type: "json" };
const readPackageJson = async () => {
    const packageJsonPath = path.join(".", "package.json");
    const contents = await fs.promises.readFile(packageJsonPath);
    return JSON.parse(contents);
};

const parsePackageJson = (packageJson) => {
    const arch = ARCH_MAPPING[process.arch];
    if (!arch) {
        throw Error(
            `Installation is not supported for this architecture: ${process.arch}`
        );
    }

    const platform = PLATFORM_MAPPING[process.platform];
    if (!platform) {
        throw Error(
            `Installation is not supported for this platform: ${process.platform}`
        );
    }

    // Build the download url from package.json
    const version = packageJson.version;
    const repo = packageJson.repository;
    const url = `https://github.com/${repo}/releases/download/v${version}/cli_${version}_${platform}_${arch}.tar.gz`;

    let binPath = path.join("bin", "durable");
    if (platform === "windows") {
        binPath += ".exe";
    }

    return { binPath, url };
};

const errGlobal = "Installing Durable CLI as a global module is not supported.";

/**
 * Reads the configuration from application's package.json,
 * downloads the binary from package url and stores at
 * ./bin in the package's root.
 *
 *  See: https://docs.npmjs.com/files/package.json#bin
 */
async function main() {
    const yarnGlobal = JSON.parse(
        process.env.npm_config_argv || "{}"
    ).original?.includes("global");
    if (process.env.npm_config_global || yarnGlobal) {
        throw errGlobal;
    }

    const pkg = await readPackageJson();
    const { binPath, url } = parsePackageJson(pkg);
    const binDir = path.dirname(binPath);
    await fs.promises.mkdir(binDir, { recursive: true });


    console.info("Downloading", url );
    const res = await fetch(url);
    const fileStream = fs.createWriteStream(binPath);

    await new Promise((resolve, reject) => {
        res.body.pipe(fileStream);
        res.body.on("error", reject);
        fileStream.on("finish", resolve);
    });
    // Link the binaries in postinstall to support yarn
    await binLinks({
        path: path.resolve("."),
        pkg: { ...pkg, bin: { [pkg.name]: binPath } },
    });
   


    // TODO: verify checksums
    console.info("Installed Durable CLI successfully");
}

await main();