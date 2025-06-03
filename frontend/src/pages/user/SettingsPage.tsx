import React, { useState } from 'react';
import { Layout } from '../../components/layout/Layout';
import { useAuth } from '../../hooks/useAuth';
import { mockCredits, mockCreditPackages, mockTransactionHistory, mockUserSettings } from '../../utils/mockData';
import '../../styles/pages/settings.css';

type SettingSection = 'account' | 'credits' | 'notifications' | 'privacy';

export const SettingsPage: React.FC = () => {
  const { user } = useAuth();
  const [activeSection, setActiveSection] = useState<SettingSection>('account');
  const [formData, setFormData] = useState({
    displayName: user?.username || 'Alex Johnson',
    username: user?.username || 'alex_johnson',
    email: user?.email || 'alex@example.com'
  });
  const [notifications, setNotifications] = useState({
    worldGeneration: true,
    newFeatures: true,
    creditLow: false,
    marketing: false
  });
  const [privacy, setPrivacy] = useState({
    analytics: true
  });

  const handleInputChange = (field: string, value: string) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleNotificationToggle = (setting: string) => {
    setNotifications(prev => ({
      ...prev,
      [setting]: !prev[setting as keyof typeof prev]
    }));
  };

  const handlePrivacyToggle = (setting: string) => {
    setPrivacy(prev => ({
      ...prev,
      [setting]: !prev[setting as keyof typeof prev]
    }));
  };

  const handleSaveAccount = () => {
    console.log('Saving account settings:', formData);
    // TODO: Implement actual API call
    alert('Account settings saved!');
  };

  const handleSaveNotifications = () => {
    console.log('Saving notification settings:', notifications);
    // TODO: Implement actual API call
    alert('Notification preferences saved!');
  };

  const handleSavePrivacy = () => {
    console.log('Saving privacy settings:', privacy);
    // TODO: Implement actual API call
    alert('Privacy settings saved!');
  };

  const handleBuyCredits = (amount: number, price: number) => {
    console.log(`Purchasing ${amount} credits for $${price}`);
    // TODO: Implement actual payment processing
    alert(`Purchasing ${amount} credits for $${price.toFixed(2)}`);
  };

  const handleChangePassword = () => {
    // TODO: Implement password change modal
    alert('Password change functionality coming soon!');
  };

  const handleChangePicture = () => {
    // TODO: Implement picture upload functionality
    alert('Picture change functionality coming soon!');
  };

  const handleDownloadData = () => {
    console.log('Downloading user data...');
    // TODO: Implement data export
    alert('Data export functionality coming soon!');
  };

  const handleDeleteAccount = () => {
    const confirmed = window.confirm(
      'Are you sure you want to delete your account? This action cannot be undone.'
    );
    if (confirmed) {
      console.log('Deleting account...');
      // TODO: Implement account deletion
      alert('Account deletion functionality coming soon!');
    }
  };

  const NavItem: React.FC<{ section: SettingSection; children: React.ReactNode }> = ({ section, children }) => (
    <button
      onClick={() => setActiveSection(section)}
      className={`nav-item ${activeSection === section ? 'active' : ''}`}
    >
      {children}
    </button>
  );

  const ToggleSwitch: React.FC<{ 
    checked: boolean; 
    onChange: () => void;
    label: string;
    description: string;
  }> = ({ checked, onChange, label, description }) => (
    <div className="setting-item">
      <div className="setting-info">
        <h3>{label}</h3>
        <p>{description}</p>
      </div>
      <label className="toggle-switch">
        <input
          type="checkbox"
          checked={checked}
          onChange={onChange}
          className="toggle-input"
        />
        <span className="toggle-slider"></span>
      </label>
    </div>
  );

  return (
    <Layout>
      <div className="min-h-screen flex flex-col bg-white">
        <main className="settings-page">
          <div className="settings-container">
            
            {/* PAGE HEADER */}
            <div className="settings-header">
              <h1 className="settings-title">Settings</h1>
              <p className="settings-subtitle">Manage your account, credits, and preferences.</p>
            </div>

            <div className="settings-layout">
              
              {/* SIDEBAR NAVIGATION */}
              <div className="settings-sidebar">
                <nav className="settings-nav">
                  <NavItem section="account">Account Settings</NavItem>
                  <NavItem section="credits">Credits & Billing</NavItem>
                  <NavItem section="notifications">Notifications</NavItem>
                  <NavItem section="privacy">Privacy</NavItem>
                </nav>
              </div>

              {/* MAIN SETTINGS CONTENT */}
              <div className="settings-content">
                
                {/* ACCOUNT SETTINGS */}
                {activeSection === 'account' && (
                  <div className="section-content">
                    
                    {/* Profile Information */}
                    <div className="settings-card">
                      <h2 className="card-title">Profile Information</h2>
                      
                      {/* Profile Picture */}
                      <div className="profile-picture-section">
                        <div 
                          className="profile-avatar" 
                          style={{ 
                            backgroundImage: `url('/no-image.jpg')` 
                          }}
                        ></div>
                        <div className="profile-avatar-info">
                          <button 
                            className="btn btn-primary"
                            onClick={handleChangePicture}
                          >
                            Change Picture
                          </button>
                          <p>JPG, PNG up to 5MB</p>
                        </div>
                      </div>

                      <div className="form-group">
                        <label className="form-label">Display Name</label>
                        <input 
                          type="text" 
                          value={formData.displayName}
                          onChange={(e) => handleInputChange('displayName', e.target.value)}
                          className="form-input"
                        />
                      </div>
                      <div className="form-group">
                        <label className="form-label">Username</label>
                        <input 
                          type="text" 
                          value={formData.username}
                          onChange={(e) => handleInputChange('username', e.target.value)}
                          className="form-input"
                        />
                      </div>
                    </div>

                    {/* Account Security */}
                    <div className="settings-card">
                      <h2 className="card-title">Account Security</h2>
                      <div className="form-group">
                        <label className="form-label">Email Address</label>
                        <input 
                          type="email" 
                          value={formData.email}
                          onChange={(e) => handleInputChange('email', e.target.value)}
                          className="form-input"
                        />
                      </div>
                      <div className="form-group">
                        <label className="form-label">Password</label>
                        <button 
                          className="btn btn-secondary"
                          onClick={handleChangePassword}
                        >
                          Change Password
                        </button>
                      </div>
                      
                      <div className="section-divider">
                        <button 
                          className="btn btn-primary"
                          onClick={handleSaveAccount}
                        >
                          Save Changes
                        </button>
                      </div>
                    </div>

                  </div>
                )}

                {/* CREDITS & BILLING */}
                {activeSection === 'credits' && (
                  <div className="section-content">
                    
                    {/* Current Balance */}
                    <div className="settings-card">
                      <h2 className="card-title">Credit Balance</h2>
                      <div className="credits-balance">
                        <div className="balance-info">
                          <div className="balance-amount">{mockCredits.balance.toLocaleString()} 💎</div>
                          <p className="balance-label">Available Credits</p>
                        </div>
                        <button className="btn btn-primary btn-lg">Buy More Credits</button>
                      </div>
                      
                      {/* Credit Packages */}
                      <div className="credit-packages">
                        {mockCreditPackages.map((pkg) => (
                          <div 
                            key={pkg.id}
                            className={`credit-package ${pkg.popular ? 'popular' : ''}`}
                            onClick={() => handleBuyCredits(pkg.credits, pkg.price)}
                          >
                            <div className="package-amount">{pkg.credits} 💎</div>
                            <div className="package-price">${pkg.price}</div>
                            {pkg.popular && <div className="package-badge">Most Popular</div>}
                            <button className="package-button">Buy</button>
                          </div>
                        ))}
                      </div>
                    </div>

                    {/* Transaction History */}
                    <div className="settings-card">
                      <h2 className="card-title">Transaction History</h2>
                      <div className="transaction-list">
                        {mockTransactionHistory.map((transaction) => (
                          <div key={transaction.id} className="transaction-item">
                            <div className="transaction-info">
                              <h3>{transaction.description}</h3>
                              <p>{new Date(transaction.date).toLocaleDateString()} • 14:23</p>
                            </div>
                            <span className={`transaction-amount ${transaction.amount > 0 ? 'positive' : 'negative'}`}>
                              {transaction.amount > 0 ? '+' : ''}{transaction.amount} 💎
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>

                  </div>
                )}

                {/* NOTIFICATIONS */}
                {activeSection === 'notifications' && (
                  <div className="section-content">
                    <div className="settings-card">
                      <h2 className="card-title">Notification Preferences</h2>
                      
                      <ToggleSwitch
                        checked={notifications.worldGeneration}
                        onChange={() => handleNotificationToggle('worldGeneration')}
                        label="World Generation Complete"
                        description="Get notified when your world generation is finished"
                      />

                      <ToggleSwitch
                        checked={notifications.newFeatures}
                        onChange={() => handleNotificationToggle('newFeatures')}
                        label="New World Features"
                        description="Updates about new features and improvements"
                      />

                      <ToggleSwitch
                        checked={notifications.creditLow}
                        onChange={() => handleNotificationToggle('creditLow')}
                        label="Credit Balance Low"
                        description="Alert when your credit balance is running low"
                      />

                      <ToggleSwitch
                        checked={notifications.marketing}
                        onChange={() => handleNotificationToggle('marketing')}
                        label="Marketing Emails"
                        description="Promotional content and special offers"
                      />
                      
                      <div className="section-divider">
                        <button 
                          className="btn btn-primary"
                          onClick={handleSaveNotifications}
                        >
                          Save Preferences
                        </button>
                      </div>
                    </div>
                  </div>
                )}

                {/* PRIVACY */}
                {activeSection === 'privacy' && (
                  <div className="section-content">
                    <div className="settings-card">
                      <h2 className="card-title">Privacy Settings</h2>
                      
                      <div className="privacy-section">
                        <h3>Data Usage</h3>
                        <ToggleSwitch
                          checked={privacy.analytics}
                          onChange={() => handlePrivacyToggle('analytics')}
                          label="Analytics & Performance"
                          description="Help improve our service by sharing usage data"
                        />
                      </div>

                      <div className="privacy-section">
                        <h3>Account Management</h3>
                        <div className="account-management">
                          <button 
                            className="btn btn-secondary"
                            onClick={handleDownloadData}
                          >
                            Download My Data
                          </button>
                          <button 
                            className="btn danger"
                            onClick={handleDeleteAccount}
                          >
                            Delete Account
                          </button>
                        </div>
                      </div>
                      
                      <div className="section-divider">
                        <button 
                          className="btn btn-primary"
                          onClick={handleSavePrivacy}
                        >
                          Save Settings
                        </button>
                      </div>
                    </div>
                  </div>
                )}

              </div>
            </div>

          </div>
        </main>
      </div>
    </Layout>
  );
};