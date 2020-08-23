## Intro

Allows you to show your latest douban activities on your github profile or project readme

## Usage:

### Example Workflow file

```yaml
uses: dongweimng/douban-activity-readme@master
with:
  uid: YOUR_DOUBAN_USER_ID
  max_count: 5
```

### Example TEMPLATE.md file

```markdown

### 🗣 My activity:

<!-- DOUBAN-ACTIVITIES:START -->
<!-- DOUBAN-ACTIVITIES:END -->
```

##### Configuration

| option | value   | default   | description |
| ------ | ------- | --------- | ----------- |
| uid | string  | `false` | Your douban user_id. |
| max_count   | string  | `5` | Maximum number of activities you want to show on your readme, all feeds combined. |
| readme_path    | string | `./README.md`  | Path of the readme file you want to update. |
