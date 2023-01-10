# No, really, why the fuck you'd write another tool?

Here's a list of my grievances and ideas, since there's already much nicer tools in this space.

## Configuration means addressable values

- files and folder structures are great for organizing stuff, and there's already a bunch of great tools to operate on them. Configuration values can be arranged into a multitude of different _config trees_. files and folder names are less than ideal when the interface is not a filesystem, for example during IPC/RPC, but URIs are!
- configuration trees are collections of configuration _entries_:
  1. a **key**, or more likely, a path (for example: `smtp.password`).
  2. a **type**, or kind of value it holds, such as secrets, scalar values or collections.
  3. the **value**, that is, the actual data representing the configuration value, and,
  4. **comments**, let's keep happy the human in the loop!
- the _key_ of a configuration value is actually formed of two keys: a **global** and a **local** one. The global one depends on who's asking, that is, a local user may use _file names_ (`api/config.prod.yaml`), while an automated process should use an _URI_ (`op://prod/api`). The local one is always the same, regardless of who's asking, i.e. `smtp.password`. Adding these together produces a _fully qualified key_, or a pointer to the corresponding entry's value.
- So being _addressable_ means configuration _values_ can be extracted from _config trees_ by their _fully qualified key_.

For example, let's say we've got a file named `api/config.prod.yaml` with the following contents:

```yaml
_config: !!joao
  vault: prod
  name: api
smtp:
  username: alguem
  password: !!secret "hô-bá-lá-lá"
```

There's many entries in that tree, and the SMTP password value would be addressed with:

- `joao get api/config.prod.yaml smtp.password`
- `op item get op://prod/api/smtp/password`
- `vault kv read -field=smtp.password config/tree/prod/api`

## Source of truth is hard

- Running a single source of truth (SSoT) for our config values also means worrying about disaster recovery, high availability, opsec and many more wonderful things that I don't feel the least inclined to worry about. 1Password already worries about it, and that's good enough for me personally.
- I don't have a need for a SSoT, really. Since it's only me, there's no need for a change-management, auditing and reporting process. Buuuuut, there is in fact two of them: git and 1password! That doesn't mean they can prevent me from doing something stupid, or evil, but allow me to experiment without adding process I personally don't need on my personal projects.
- Automated processes running on their own need something like a SSoT, or at least a source of truth. Can't rely on git since that means automated processes (think a scheduler listening on changes) would need to have credentials (for git transport), a process to refresh, probably some keys to decrypt data, and all that jazz. That's what 1Password connect is for.
- So okay, at least a few of sources of truth:
  1. my **local filesystem**, where I experiment and provision. At some points, this can be considered _the_ source of truth, since it contains all components of a config value: the key, type, data and comments.
  2. another up at **git remotes** keeping everything but secret values, this is better than `.template.yaml` files since they're also useful to share knowledge and keep track of progress,
  3. last one in **1Password**, one item per file, sans comments
- But now you've got secrets in plaintext lying around in the local fileystem 😱! Preventing these from landing in the wrong remote filesystems is important, so scrubbing them after editing/before uploading is necessary.

## Useful for humans and robots alike

- I'd like a tool that doesn't make a tradeoff between my robots' convenience and mine. I can let Vault figure out permissions and access for robots. I can unlock 1Password locally with my fingerprint. I can use 1Password Connect tokens for provisioning new hardware or vms. Tooling needs to catch up to my personal needs.
- YAML is fine because comments and "custom" types (i.e. `!!secret`), but that's about all I like about it. I'm sure this is going to become a pain later on, but until that day arrives, I'm sticking to YAML.
- YAML sucks because it's not JSON and `jq` only does JSON. Robots love JSON, but JSON doesn't have comments and it writing it by hand makes me sad (but also, why would you when there's `jq`!). So there's `yq` you say, and it's truly great! but also not `jq` and now you need `yq` and `jq` (plus bash, because reasons) to make local scripting a thing.
- Now try editing YAML in a phone screen's default textbox editor. Yeah, it's going to suck. I can workflow my way around it but I don't wanna build and maintain UIs or sync processes. I mostly just wanna do quick edits or saves on particular fields, so losing comments is fine, really trusting dvcs to not drop the ball there. By translating entries into 1Password fields, I rely on them to do the heavy lifting.
- Robots love JSON, 1Password talks JSON (unless `op item edit` 😢), Vault talks JSON, turning YAML into JSON (and back) is pretty easy and there's golang libraries for all of this. Executable has to be good enough for robot and roborto 🤖, so it needs to deal with JSON and YAML.
- git is hard, but `git-crypt` seems to work great, need to make sure this tool makes doing something stupid harder.
