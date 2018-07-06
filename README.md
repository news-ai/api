# api

API for NewsAI platform. For products:

1. Media List Management (Tabulae)
2. Media Database

In the `/api/` folder:

- Running development: `goapp serve local.yaml`
- Deploying: `goapp deploy` (can be either `dev.yaml` or `prod.yaml`)
- Rollback: `appcfg.py rollback -A newsai-1166 -V 1 api/`

Indexes:

- Update: `appcfg.py update_indexes api/ -A newsai-1166`
- Delete: `appcfg.py vacuum_indexes api/ -A newsai-1166`

Cron:

- `appcfg.py update_cron api/ -A newsai-1166`

Clearbit risk:

```
curl https://risk.clearbit.com/v1/calculate \
        -d 'email=bitabidem@10vpn.info' \
        -d 'ip=103.85.161.6' \
        -u sk_e571cbd973ecee8874cdbc33559e7480
```
