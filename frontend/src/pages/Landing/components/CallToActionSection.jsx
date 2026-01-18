import { Link } from 'react-router-dom';

const CallToActionSection = () => {
  return (
    <section id="apply" className="cta-section">
      <div className="container">
        <div className="cta-content">
          <h2>Подать заявку на обучение</h2>
          <p>Заполните форму заявки, и наши специалисты свяжутся с вами в ближайшее время. После одобрения заявки вы получите данные для входа в систему через Telegram.</p>
          <Link to="/login" className="btn btn-primary">
            Подать заявку
          </Link>
        </div>
      </div>
    </section>
  );
};

export default CallToActionSection;
