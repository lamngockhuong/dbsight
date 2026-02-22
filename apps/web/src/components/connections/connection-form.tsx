import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  buildDsn,
  DB_TYPE_LABELS,
  type DbType,
  type DsnFields,
  getDefaultPort,
  getDsnPlaceholder,
} from '@/lib/dsn-builder'

interface ConnectionFormProps {
  onSubmit: (data: { name: string; db_type: string; dsn: string }) => Promise<void>
}

const DB_TYPES: DbType[] = ['postgres', 'mysql', 'mariadb']

export function ConnectionForm({ onSubmit }: ConnectionFormProps) {
  const [name, setName] = useState('')
  const [dbType, setDbType] = useState<DbType>('postgres')
  const [inputMode, setInputMode] = useState<'dsn' | 'form'>('dsn')
  const [dsn, setDsn] = useState('')
  const [fields, setFields] = useState<DsnFields>({
    host: 'localhost',
    port: '',
    username: '',
    password: '',
    database: '',
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const updateField = (key: keyof DsnFields, value: string) => {
    setFields((prev) => ({ ...prev, [key]: value }))
  }

  const handleDbTypeChange = (newType: DbType) => {
    setDbType(newType)
    setFields((prev) => ({ ...prev, port: getDefaultPort(newType) }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    try {
      const finalDsn = inputMode === 'form' ? buildDsn(dbType, fields) : dsn
      await onSubmit({ name, db_type: dbType, dsn: finalDsn })
      setName('')
      setDsn('')
      setFields({ host: 'localhost', port: '', username: '', password: '', database: '' })
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Add Connection</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="flex flex-col gap-3">
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Connection name"
            required
          />

          {/* Database type selector */}
          <select
            value={dbType}
            onChange={(e) => handleDbTypeChange(e.target.value as DbType)}
            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          >
            {DB_TYPES.map((t) => (
              <option key={t} value={t}>
                {DB_TYPE_LABELS[t]}
              </option>
            ))}
          </select>

          {/* Input mode toggle */}
          <div className="flex gap-1 rounded-md border p-0.5">
            <button
              type="button"
              onClick={() => setInputMode('dsn')}
              className={`flex-1 rounded px-2 py-1 text-xs font-medium transition-colors ${
                inputMode === 'dsn' ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'
              }`}
            >
              DSN String
            </button>
            <button
              type="button"
              onClick={() => setInputMode('form')}
              className={`flex-1 rounded px-2 py-1 text-xs font-medium transition-colors ${
                inputMode === 'form' ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'
              }`}
            >
              Form Fields
            </button>
          </div>

          {inputMode === 'dsn' ? (
            <Input
              type="password"
              value={dsn}
              onChange={(e) => setDsn(e.target.value)}
              placeholder={getDsnPlaceholder(dbType)}
              required
            />
          ) : (
            <div className="grid grid-cols-2 gap-2">
              <Input
                value={fields.host}
                onChange={(e) => updateField('host', e.target.value)}
                placeholder="Host"
                required
              />
              <Input
                value={fields.port}
                onChange={(e) => updateField('port', e.target.value)}
                placeholder={`Port (${getDefaultPort(dbType)})`}
              />
              <Input
                value={fields.username}
                onChange={(e) => updateField('username', e.target.value)}
                placeholder="Username"
                required
              />
              <Input
                type="password"
                value={fields.password}
                onChange={(e) => updateField('password', e.target.value)}
                placeholder="Password"
                required
              />
              <Input
                className="col-span-2"
                value={fields.database}
                onChange={(e) => updateField('database', e.target.value)}
                placeholder="Database name"
                required
              />
            </div>
          )}

          <Button type="submit" disabled={loading}>
            {loading ? 'Saving...' : 'Add Connection'}
          </Button>
          {error && <p className="text-sm text-destructive">{error}</p>}
        </form>
      </CardContent>
    </Card>
  )
}
