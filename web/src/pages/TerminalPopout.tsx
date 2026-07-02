import { useEffect } from "react";
import { useParams, useSearchParams } from "react-router-dom";
import { DeviceTerminal } from "../components/DeviceTerminal";

// TerminalPopout rendert ein einzelnes Vollfenster-Terminal ohne App-Chrome –
// Ziel des „Popout"-Buttons (window.open, gleiche Origin → Cookie-Auth gilt).
export function TerminalPopout() {
  const { id } = useParams<{ id: string }>();
  const [params] = useSearchParams();
  const os = params.get("os") || "";
  const shell = params.get("shell") || undefined;
  const runas = (params.get("runas") as "system" | "user") || undefined;

  useEffect(() => {
    const prev = document.title;
    document.title = `Terminal · ${id}`;
    return () => { document.title = prev; };
  }, [id]);

  if (!id) return null;
  return (
    <div className="term-popout">
      <DeviceTerminal id={id} os={os} fill autoStart initialShell={shell} initialRunas={runas} />
    </div>
  );
}
