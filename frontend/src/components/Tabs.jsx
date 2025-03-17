import React from 'react';

const Tabs = ({ activeTab, onTabChange, darkMode }) => {
  const tabs = [
    { id: 'tous', label: 'Tous' },
    { id: 'actifs', label: 'Actifs' },
    { id: 'queued', label: 'En attente' },
    { id: 'terminés', label: 'Terminés' },
    { id: 'pause', label: 'Pause' }
  ];

  return (
    <div className={`flex border-b ${darkMode ? 'border-gray-700' : 'border-gray-200'} px-6`}>
      {tabs.map(tab => (
        <button 
          key={tab.id}
          onClick={() => onTabChange(tab.id)}
          className={`px-5 py-3 font-medium transition-colors ${
            activeTab === tab.id 
              ? `border-b-2 border-blue-500 ${darkMode ? 'text-blue-400' : 'text-blue-600'}` 
              : `${darkMode ? 'text-gray-400 hover:text-gray-300' : 'text-gray-500 hover:text-gray-700'}`
          }`}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
};

export default Tabs;