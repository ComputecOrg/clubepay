"use client";

import { QRCodeSVG } from "qrcode.react";

interface QRCodeProps {
  slug: string;
}

export default function BusinessQRCode({ slug }: QRCodeProps) {
  const url =
    typeof window !== "undefined"
      ? `${window.location.origin}/validar/${slug}`
      : `/validar/${slug}`;

  function handlePrint() {
    const printWindow = window.open("", "_blank");
    if (!printWindow) return;
    printWindow.document.write(`
      <html>
        <head><title>QR Code - ClubePay</title></head>
        <body style="display:flex;flex-direction:column;align-items:center;justify-content:center;min-height:100vh;font-family:system-ui">
          <h2 style="color:#2a7d6e">Escaneie para validar seu uso</h2>
          <div id="qr"></div>
          <p style="margin-top:16px;color:#666">${url}</p>
          <script>window.print();</script>
        </body>
      </html>
    `);
    printWindow.document.close();
  }

  return (
    <div className="bg-white rounded-2xl border border-gray-200 p-6 flex flex-col items-center gap-4">
      <h2 className="font-semibold text-gray-800">QR Code do balcao</h2>
      <p className="text-sm text-gray-500 text-center">
        Imprima e coloque no balcao. Assinantes escaneiam para validar o uso.
      </p>
      <QRCodeSVG value={url} size={200} fgColor="#2a7d6e" />
      <p className="text-xs text-gray-400 break-all text-center">{url}</p>
      <button
        onClick={handlePrint}
        className="rounded-xl px-4 py-2 text-sm font-semibold text-white transition-opacity hover:opacity-90"
        style={{ backgroundColor: "#d4a853" }}
      >
        Imprimir QR Code
      </button>
    </div>
  );
}
