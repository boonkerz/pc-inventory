import { Navigate, Route, Routes } from "react-router-dom";
import { useAuth } from "./auth";
import { Layout } from "./components/Layout";
import { Login } from "./pages/Login";
import { Dashboard } from "./pages/Dashboard";
import { Devices } from "./pages/Devices";
import { DeviceDetail } from "./pages/DeviceDetail";
import { Policies } from "./pages/Policies";
import { Scripts } from "./pages/Scripts";
import { TwoFactorSetup } from "./pages/TwoFactorSetup";
import { TerminalPopout } from "./pages/TerminalPopout";

export default function App() {
  const { user, loading } = useAuth();

  if (loading) return <div className="center muted">Lädt…</div>;
  if (!user) return <Login />;
  // 2FA-Pflicht: ohne aktivierten zweiten Faktor zuerst die Einrichtung erzwingen.
  if (user.require_2fa && !user.totp_enabled) return <TwoFactorSetup />;

  return (
    <Routes>
      {/* Popout-Terminal: eigenes Vollfenster ohne Layout/Sidebar. */}
      <Route path="/devices/:id/terminal" element={<TerminalPopout />} />
      <Route path="*" element={
        <Layout>
          <Routes>
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/devices" element={<Devices />} />
            <Route path="/devices/:id" element={<DeviceDetail />} />
            <Route path="/policies" element={<Policies />} />
            <Route path="/scripts" element={<Scripts />} />
            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </Layout>
      } />
    </Routes>
  );
}
