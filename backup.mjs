#!/usr/bin/env zx

// import "zx/globals";

async function loadConfig(path) {
  let content = {};
  let isNewConfig = true;
  if (fs.pathExistsSync(path)) {
    isNewConfig = false;
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
  if (isNewConfig) {
    await fs.writeFile(path, JSON.stringify(content, null, 2));
  }
  return content;
}

async function fetchRepos(username, token) {
  let store = {};
  let page = 0;

  while (true) {
    let response = await fetch(
      `https://api.github.com/search/repositories?q=user%3A${encodeURIComponent(username)}&page=${page}`,
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

String.prototype.splitOnce = function (ele) {
  var i = this.indexOf(ele);
  return [this.slice(0, i), this.slice(i + 1)];
};

function parseArgs() {
  let res = {};
  for (const c of process.argv.filter((arg) => arg.startsWith("--"))) {
    const [key, value] = c.splitOnce("=");
    res[key.replace("--", "")] = value;
  }

  if (res.target) {
    res.target = path.resolve(res.target);
    cd(res.target);
  } else {
    res.target = path.resolve(".");
  }

  if (res.config) {
    res.config = path.resolve(res.config);
  } else {
    res.config = path.resolve("./.github_backup_config.json");
  }
  return res;
}

const yesOrNoChoices = { choices: ["y", "Y", "n", "N"] };
const yesOrNoToBoolean = { y: true, n: false, Y: true, N: false };

const args = parseArgs();

// load config
let config = await loadConfig(args.config);

// clone without question
const cloneAll = args.clone === "all";

// fetch repos
const keepRepos = {};
const ignoreRepos = {};
const remoteRepos = await fetchRepos(config.username, config.token);
const remoteReposKeys = Object.keys(remoteRepos);

// handle untracked repositories
for (const name of Object.keys(config.repos)) {
  const repo = config.repos[name];
  const repoDir = path.resolve(`./${repo.name}`);
  if (remoteReposKeys.includes(repo.name)) {
    remoteRepos[repo.name].ignore = repo.ignore;
    // ignore repositories if need
    if (repo.ignore) {
      ignoreRepos[repo.name] = repo;
      delete remoteRepos[repo.name];
    }
    continue;
  }
  // delete or keep untracked repositories
  if (fs.pathExistsSync(repoDir)) {
    if (repo.keep) {
      continue;
    }
    const del = await question(`Delete ${repoDir}? (y/n): `, yesOrNoChoices);
    if (yesOrNoToBoolean(del) || false) {
      await fs.remove(repoDir);
    } else {
      const keep = await question(`Keep ${repoDir}? (y/n): `, yesOrNoChoices);
      if (yesOrNoToBoolean(keep) || false) {
        repo.keep = true;
        keepRepos[repo.name] = repo;
      }
    }
  }
}

// update repos
for (const name of Object.keys(remoteRepos)) {
  const repo = remoteRepos[name];
  const repoDir = path.resolve(`./${repo.name}`);

  // clone if not exist
  if (!fs.pathExistsSync(repoDir)) {
    if (cloneAll) {
      await $`git clone ${repo.ssh_url}`;
    } else {
      const clone = await question(
        `Clone ${repo.ssh_url}? (y/n): `,
        yesOrNoChoices
      );
      if (yesOrNoToBoolean(clone) || false) {
        await $`git clone ${repo.ssh_url}`;
      } else {
        repo.ignore = true;
        continue;
      }
    }
  }

  // fetch all branch
  cd(repoDir);
  try {
    let branchs = await quiet($`git branch -r`);
    branchs = branchs.stdout
      .split("\n")
      .map((r) => r.replace(/^ */, ""))
      .filter((r) => r.indexOf("->") < 0 && r.length > 0)
      .map((r) => r.splitOnce("/"))
      .map((r) => {
        const [remote, branch] = r;
        return { remote, branch };
      });
    if (branchs.length > 0) {
      await $`git checkout --quiet --detach HEAD`;
      for (const b of branchs) {
        try {
          await $`git fetch ${b.remote} ${b.branch}`;
        } catch (p) {
          console.log(`Error: ${p.stderr || p}`);
        }
      }
      await $`git checkout --quiet -`;
    }
  } catch (p) {
    console.log(`Error: ${p.stderr || p}`);
  }
  cd(args.target);
}

// update config
config.repos = { ...keepRepos, ...ignoreRepos, ...remoteRepos };
await fs.writeFile(args.config, JSON.stringify(config, null, 2));
