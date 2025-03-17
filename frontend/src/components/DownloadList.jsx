import React from 'react';
import { FileText } from 'lucide-react';
import DownloadItem from './DownloadItem';

const DownloadList = ({ downloads, isLoading, darkMode, onDownloadAction, onRemoveDownload }) => {
  if (isLoading) {
    return (
      <div className="flex-1 overflow-auto p-6">
        <div className="flex justify-center items-center h-40">
          <div className={`animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 ${darkMode ? 'border-blue-400' : 'border-blue-500'}`}></div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-auto p-6">
      <div className="space-y-4 max-w-4xl mx-auto">
        {!downloads || downloads.length === 0 ? (
          <div className={`text-center py-16 ${darkMode ? 'text-gray-400' : 'text-gray-500'} flex flex-col items-center`}>
            <FileText size={48} className="mb-4 opacity-50" />
            <p className="text-lg">Aucun téléchargement à afficher</p>
            <p className="text-sm mt-2">Cliquez sur "Nouveau" pour ajouter un téléchargement</p>
          </div>
        ) : (
          downloads.map(download => (
            <DownloadItem 
              key={download.ID}
              download={download}
              darkMode={darkMode}
              onDownloadAction={onDownloadAction}
              onRemoveDownload={onRemoveDownload}
            />
          ))
        )}
      </div>
    </div>
  );
};

export default DownloadList;