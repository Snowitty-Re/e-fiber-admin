# atlas.hcl —— Atlas 迁移配置
# 禁止 AI 自行修改，除非用户明确要求。

variable "db_url" {
  type    = string
  default = "postgres://efa:efa_secret@localhost:5432/efa?sslmode=disable&search_path=public"
}

env "local" {
  src = "ent://internal/ent/schema"
  url = var.db_url
  dev = "docker://postgres/16/dev"
  migration {
    dir = "file://migrations"
    format = "atlas"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
      apply = "{{ sql . \"  \" }}"
    }
  }
}
