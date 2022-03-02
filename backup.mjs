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
    let response = await fetch(`https://api.github.com/search/repositories?q=user%3Atbxark&page=${page}`, {
        method: "GET",
        headers: { Authorization: `token ${token}` },
    });
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

const args = parseArgs();

// repos store directory
// let targetDir = await quiet($`pwd`);
// targetDir = targetDir.stdout.toString();
if (args["target"]) {
  // targetDir = args["target"];
  cd(args["target"]);
}

// load config
let cPath = args["config"] || "./.github_backup_config.json";
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

  // ignore repo
  if (config.repos[name] && config.repos[name].ignore) {
    remoteRepos[name].ignore = true;
    continue;
  }

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
      } else if (clone === "n") {
        remoteRepos[name].ignore = true;
        continue;
      }
    }
  }

  // fetch all branch
  cd(path);
  try {
    let branchs = await quiet($`git branch -r`);
    branchs = branchs.stdout
      .split("\n")
      .map((r) => r.replace(/^ */, ""))
      .filter((r) => r.indexOf("->") < 0 && r.length > 0)
      .map((r) => r.split("/"))
      .map((r) => {
        const [remote, ...branch] = r;
        return branch.join("/");
      })
      .map((r) => `${r}:${r}`)
      .join(" ");
    if (branchs.length > 0) {
      await $`git checkout --quiet --detach HEAD`;
      await $`git fetch origin ${branchs}`;
      await $`git checkout --quiet -`;
    }
  } catch (p) {
    console.log(`Error: ${p.stderr || p}`);
  }
  cd("..");
}

// update config
config.repos = remoteRepos;
await fs.writeFile(cPath, JSON.stringify(config, null, 2));
