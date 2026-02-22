export type DbType = 'postgres' | 'mysql' | 'mariadb'

export interface DsnFields {
  host: string
  port: string
  username: string
  password: string
  database: string
}

export function buildDsn(dbType: DbType, fields: DsnFields): string {
  const { host, port, username, password, database } = fields
  const encUser = encodeURIComponent(username)
  const encPass = encodeURIComponent(password)
  switch (dbType) {
    case 'postgres':
      return `postgres://${encUser}:${encPass}@${host}:${port || '5432'}/${database}`
    case 'mysql':
    case 'mariadb':
      return `${encUser}:${encPass}@tcp(${host}:${port || '3306'})/${database}?parseTime=true&timeout=10s`
  }
}

export function getDefaultPort(dbType: DbType): string {
  return dbType === 'postgres' ? '5432' : '3306'
}

export function getDsnPlaceholder(dbType: DbType): string {
  switch (dbType) {
    case 'postgres':
      return 'postgres://user:pass@host:5432/db'
    case 'mysql':
    case 'mariadb':
      return 'user:pass@tcp(host:3306)/db?parseTime=true'
  }
}

export const DB_TYPE_LABELS: Record<DbType, string> = {
  postgres: 'PostgreSQL',
  mysql: 'MySQL',
  mariadb: 'MariaDB',
}
