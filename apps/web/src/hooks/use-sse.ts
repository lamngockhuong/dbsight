import { useCallback, useEffect, useRef, useState } from 'react'

export function useSSE<T>(url: string | null) {
  const [data, setData] = useState<T | null>(null)
  const esRef = useRef<EventSource | null>(null)
  const urlRef = useRef(url)
  urlRef.current = url

  const connect = useCallback(() => {
    const currentUrl = urlRef.current
    if (!currentUrl) return

    const es = new EventSource(currentUrl)
    esRef.current = es

    es.onmessage = (e) => {
      try {
        setData(JSON.parse(e.data) as T)
      } catch {
        // ignore parse errors
      }
    }

    es.onerror = () => {
      es.close()
      // Reconnect after 5s if url hasn't changed
      setTimeout(() => {
        if (urlRef.current === currentUrl && esRef.current === es) {
          esRef.current = null
          connect()
        }
      }, 5000)
    }
  }, [])

  useEffect(() => {
    if (!url) return
    connect()
    return () => {
      if (esRef.current) {
        esRef.current.close()
        esRef.current = null
      }
    }
  }, [url, connect])

  return data
}
