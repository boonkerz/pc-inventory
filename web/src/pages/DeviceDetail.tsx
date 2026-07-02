import { useNavigate, useParams } from "react-router-dom";
import { DevicePanel } from "../components/DevicePanel";
import { useI18n } from "../i18n";

// Vollseiten-Ansicht eines Geräts (Direktlink / Popout). Der eigentliche Inhalt
// steckt in DevicePanel, das auch im unteren Panel der Geräteliste verwendet wird.
export function DeviceDetail() {
  const { t } = useI18n();
  const { id } = useParams();
  const nav = useNavigate();
  if (!id) return null;
  return (
    <div className="page" style={{ maxWidth: "none" }}>
      <button className="link-back" onClick={() => nav("/devices")}>← {t("Geräte")}</button>
      <DevicePanel id={id} />
    </div>
  );
}
