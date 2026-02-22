import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'

interface ConnectionFormProps {
  onSubmit: (data: { name: string; db_type: string; dsn: string }) => Promise<void>
}

export function ConnectionForm({ onSubmit }: ConnectionFormProps) {
  const [name, setName] = useState('')
  const [dsn, setDsn] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    try {
      await onSubmit({ name, db_type: 'postgres', dsn })
      setName('')
      setDsn('')
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
          <Input
            type="password"
            value={dsn}
            onChange={(e) => setDsn(e.target.value)}
            placeholder="postgres://user:pass@host:5432/db"
            required
          />
          <Button type="submit" disabled={loading}>
            {loading ? 'Saving...' : 'Add Connection'}
          </Button>
          {error && <p className="text-sm text-destructive">{error}</p>}
        </form>
      </CardContent>
    </Card>
  )
}
