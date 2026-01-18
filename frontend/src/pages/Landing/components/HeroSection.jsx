import { Link } from 'react-router-dom';

const HeroSection = () => {
  return (
    <>
      <header className="landing-header">
        <div className="container header-container">
          <div className="logo">THE BOT</div>
          <Link to="/login" className="login-link">Войти</Link>
        </div>
      </header>
      <section className="hero-section">
        <img
          src="/images/waves_big.svg"
          alt=""
          className="hero-waves"
          aria-hidden="true"
        />
        <div className="hero-content">
          <div className="hero-logo">THE BOT</div>
          <h1 className="hero-title">ТВОЙ КРАТЧАЙШИЙ ПУТЬ К УСПЕХУ</h1>
        </div>
      </section>
    </>
  );
};

export default HeroSection;
