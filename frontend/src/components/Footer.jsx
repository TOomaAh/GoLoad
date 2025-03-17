import React from 'react';
import { RefreshCw } from 'lucide-react';
import { formatSpeed } from '../utils/formatters';

const Footer = ({ downloads, darkMode, onRefresh }) => {
  const activeDownloads = downloads ? downloads.filter(d => d.status === 'downloading') : [];
  
  // Calculer la vitesse globale
  const totalSpeed = activeDownloads.reduce((total, download) => total + (download.speed || 0), 0);

  return (
    <footer className={`${darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'} border-t p-4`}>
      <div className="flex justify-between items-center max-w-4xl mx-auto text-sm">
        <div>
          {activeDownloads.length} téléchargements actifs
        </div>
        <div className="flex items-center">
          <span className="mr-3">
            Vitesse globale: {formatSpeed(totalSpeed)}
          </span>
          <button 
            onClick={onRefresh} 
            className={`p-1 rounded-md ${darkMode ? 'hover:bg-gray-700' : 'hover:bg-gray-100'} transition-colors`}
          >
            <RefreshCw size={16} />
          </button>
        </div>
      </div>
    </footer>
  );
};

export default Footer;