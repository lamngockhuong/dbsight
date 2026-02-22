import { BrowserRouter, Route, Routes, useParams } from 'react-router-dom'
import { Layout } from '@/components/layout/layout'
import { ConnectionsPage } from '@/pages/connections-page'
import { DashboardPage } from '@/pages/dashboard-page'
import { ExplainPage } from '@/pages/explain-page'
import { IndexesPage } from '@/pages/indexes-page'
import { PastePage } from '@/pages/paste-page'
import { QueriesPage } from '@/pages/queries-page'

function ConnLayout({ children }: { children: React.ReactNode }) {
  const { id } = useParams()
  return <Layout connectionId={id ? parseInt(id, 10) : undefined}>{children}</Layout>
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route
          path="/"
          element={
            <Layout>
              <DashboardPage />
            </Layout>
          }
        />
        <Route
          path="/connections"
          element={
            <Layout>
              <ConnectionsPage />
            </Layout>
          }
        />
        <Route
          path="/queries/:id"
          element={
            <ConnLayout>
              <QueriesPage />
            </ConnLayout>
          }
        />
        <Route
          path="/explain/:id"
          element={
            <ConnLayout>
              <ExplainPage />
            </ConnLayout>
          }
        />
        <Route
          path="/indexes/:id"
          element={
            <ConnLayout>
              <IndexesPage />
            </ConnLayout>
          }
        />
        <Route
          path="/paste"
          element={
            <Layout>
              <PastePage />
            </Layout>
          }
        />
      </Routes>
    </BrowserRouter>
  )
}
