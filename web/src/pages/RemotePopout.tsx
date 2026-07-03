import { useEffect } from "react";
import { useParams, useSearchParams } from "react-router-dom";
import { DeviceRemote } from "../components/DeviceRemote";

// RemotePopout rendert eine Vollfenster-Fernsteuerung ohne App-Chrome – Ziel des
// „Popout"-Buttons (window.open, gleiche Origin → Cookie-Auth gilt).
export function RemotePopout() {
  const { id } = useParams<{ id: string }>();
  const [params] = useSearchParams();
  const os = params.get("os") || "";

  useEffect(() => {
    const prev = document.title;
    document.title = `Fernsteuerung · ${id}`;
    return () => { document.title = prev; };
  }, [id]);

  if (!id) return null;
  return (
    <div className="remote-popout">
      <DeviceRemote id={id} os={os} fill autoStart />
    </div>
  );
}
