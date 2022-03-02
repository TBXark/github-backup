#!/usr/bin/env zx

//import "zx/globals";

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
  if (!content.repos)  {
    content.repos = {}
  }
  return content;
}

async function fetchRepos(username, token) {
  let response = await fetch(`https://api.github.com/users/${username}/repos`, {
    method: "GET",
    headers: { Authorization: `token ${token}` },
  });
  response = await response.json();
  response = response.map((ele) => {
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
  let store = {};
  for (const repo of response) {
    store[repo.name] = repo;
  }
  return store;
}

function updateRepos(config, remoteRepos) {
  const localReposKeys = Object.keys(config.repos)
  config.repos = remoteRepos
  return Object.keys(remoteRepos).filter(r => !localReposKeys.includes(r.name))
}

function configPath() {
  let path = process.argv.filter(x => x.startsWith('--path=')).map(x => x.replace('--path=', ''))
  if (path.length == 0) {
    return './.github_backup_config.json'
  } else {
    return path[0]
  }
}

let cPath = configPath()
let config = await loadConfig(cPath);
const remoteRepos = await fetchRepos(config.username, config.token)

for (const name of updateRepos(config, remoteRepos)) {
  const path = `./${name}`
  if (fs.pathExistsSync(path)) {
    const del = await question(`Delete ${path}? (y/n): `, {
      choices: ['y', 'n']
    })
    if (del === 'y') {
      await fs.remove(path)
    }
  }
}

let tasks = []
for (const name of Object.keys(config.repos)) {
  const path = `./${name}`
  const repo = remoteRepos[name]
  if (fs.pathExistsSync(path)) {
    tasks.push(`cd ${path} && git pull`)
  } else {
    const clone = await question(`Clone ${repo.ssh_url}? (y/n): `, {
      choices: ['y', 'n']
    })
    if (clone === 'y') {
      tasks.push(`git clone ${repo.ssh_url}`)
    }
  }
}

for (const task of tasks) {
  await `${task}`
}

await fs.writeFile(cPath, JSON.stringify(config, null, 2))
