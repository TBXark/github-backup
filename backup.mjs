#!/usr/bin/env zx

// import "zx/globals";

async function loadConfig(path) {
  let content = {};
  if (fs.pathExistsSync(path)) {
    content = await fs.readJson(path);
  }
  if (!content.username) {
    content.username = await question("Please enter your username: ");
  }
  if (!content.token) {
    content.token = await question("Please enter your token: ");
  }
  if (!content.repos) {
    content.repos = {};
  }
  return content;
}

async function fetchRepos(username, token) {
  let store = {};
  let page = 0;

  while (true) {
    let response = await fetch(
      `https://api.github.com/search/repositories?q=user%3Atbxark&page=${page}`,
      {
        method: "GET",
        headers: { Authorization: `token ${token}` },
      }
    );
    response = await response.json();
    const total = response.total_count;
    response = response.items.map((ele) => {
      return {
        name: ele.name,
        status: {
          private: ele.private,
          fork: ele.fork,
          archived: ele.archived,
        },
        date: {
          created_at: ele.created_at,
          updated_at: ele.updated_at,
        },
        ssh_url: ele.ssh_url,
      };
    });
    if (response.length == 0) {
      break;
    }
    for (const repo of response) {
      store[repo.name] = repo;
    }
    page += 1;
    if (Object.keys(store).length >= total) {
      break;
    }
  }

  return store;
}

function parseArgs() {
  let res = {};
  for (const c of process.argv.filter((arg) => arg.startsWith("--"))) {
    const [key, ...value] = c.split("=");
    res[key.replace("--", "")] = value.join("=");
  }
  return res;
}

const args = parseArgs(process.argv);

// repos store directory
let targetDir = await $`pwd`;
targetDir = targetDir.stdout;
if (args["target"]) {
  targetDir = args["target"];
  cd(args["target"]);
}

// config path
let cPath = args["config"] || "./.github_backup_config.json";

// load config
let config = await loadConfig(cPath);

// clone without question
const cloneAll = args["clone"] === "all";

// fetch repos
const remoteRepos = await fetchRepos(config.username, config.token);

const localReposKeys = Object.keys(config.repos);
const remoteReposKeys = Object.keys(remoteRepos);

// delete repos
for (const name of localReposKeys.filter((r) => !remoteReposKeys.includes(r))) {
  const path = `./${name}`;
  if (fs.pathExistsSync(path)) {
    const del = await question(`Delete ${path}? (y/n): `, {
      choices: ["y", "n"],
    });
    if (del === "y") {
      await fs.remove(path);
    }
  }
}

// update repos
for (const name of remoteReposKeys) {
  const path = `./${name}`;
  const repo = remoteRepos[name];

  // clone if not exist
  if (!fs.pathExistsSync(path)) {
    if (cloneAll) {
      await $`git clone ${repo.ssh_url}`;
    } else {
      const clone = await question(`Clone ${repo.ssh_url}? (y/n): `, {
        choices: ["y", "n"],
      });
      if (clone === "y") {
        await $`git clone ${repo.ssh_url}`;
      }
    }
  }

  // pull all branch
  cd(path);
  let branchs = await $`git branch -r`;
  branchs = branchs
    .toString()
    .split("\n")
    .map((r) => r.replace(/^ */, ""))
    .filter((r) => !r.startsWith("origin/HEAD") && r.length > 0)
    .map((r) => r.split("/"));

  for (const b of branchs) {
    const [remote, branch] = b;
    await $`git pull ${remote} ${branch}`;
  }
  cd(targetDir)
}

// update config
config.repos = remoteRepos;
await fs.writeFile(cPath, JSON.stringify(config, null, 2));
