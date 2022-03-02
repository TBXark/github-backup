#!/usr/bin/env zx

//import "zx/globals";

await $`git --version`;
await $`pwd`;

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
    const total = response.total_count
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


function configPath() {
  let path = process.argv
    .filter((x) => x.startsWith("--path="))
    .map((x) => x.replace("--path=", ""));
  if (path.length == 0) {
    return "./.github_backup_config.json";
  } else {
    return path[0];
  }
}

let cPath = configPath();
let config = await loadConfig(cPath);
const cloneAll = process.argv.indexOf("--clone=all") >= 0;
const remoteRepos = await fetchRepos(config.username, config.token);

const localReposKeys = Object.keys(config.repos);
const remoteReposKeys = Object.keys(remoteRepos);

for (const name of localReposKeys.filter(r => !remoteReposKeys.includes(r))) {
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

for (const name of remoteReposKeys) {
  const path = `./${name}`;
  const repo = remoteRepos[name];
  if (fs.pathExistsSync(path)) {
    await $`cd ${path} && git pull`;
  } else {
    if (cloneAll) {
      await $`git clone ${repo.ssh_url}`;
      continue;
    }
    const clone = await question(`Clone ${repo.ssh_url}? (y/n): `, {
      choices: ["y", "n"],
    });
    if (clone === "y") {
      await $`git clone ${repo.ssh_url}`;
    }
  }
}

config.repos = remoteRepos;
await fs.writeFile(cPath, JSON.stringify(config, null, 2));
