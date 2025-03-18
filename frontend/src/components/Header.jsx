import React from 'react';
import { Download, Plus, Settings } from 'lucide-react';

const Header = ({ darkMode, onNewDownloadClick, onSettingsClick }) => {
  return (
    <header className={`${darkMode ? 'bg-gray-800' : 'bg-white'} shadow-md px-6 py-4`}>
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold flex items-center">
          <Download className="mr-3" size={28} />
          GoLoad
        </h1>
        <div className="flex space-x-3">
          <button 
            onClick={onNewDownloadClick}
            className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors shadow-sm"
          >
            <Plus className="mr-2" size={18} />
            Nouveau
          </button>
          <button 
            onClick={onSettingsClick}
            className={`p-2 rounded-md ${darkMode ? 'bg-gray-700 text-gray-200' : 'bg-gray-200 text-gray-700'} hover:opacity-80 transition-opacity`}
          >
            <Settings size={22} />
          </button>
        </div>
      </div>
    </header>
  );
};

export default Header;