// app.js - Основний файл API сервера

const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const morgan = require('morgan');
const { Gateway, Wallets } = require('fabric-network');
const fs = require('fs');
const path = require('path');

// Ініціалізація Express
const app = express();
const port = process.env.PORT || 3000;

// Middleware
app.use(cors());
app.use(helmet());
app.use(morgan('combined'));
app.use(express.json());

// Функція для підключення до мережі
async function connectToNetwork() {
    try {
        // Шлях до гаманця користувача (адміністратора)
        const walletPath = path.join(__dirname, 'wallet');
        const wallet = await Wallets.newFileSystemWallet(walletPath);
        
        // Перевірка наявності ідентичності адміністратора в гаманці
        const identity = await wallet.get('admin');
        if (!identity) {
            console.log('Ідентичність адміністратора не знайдена в гаманці');
            return null;
        }
        
        // Шлях до профілю підключення
        const connectionProfilePath = path.join(__dirname, 'connection-org1.json');
        const connectionProfile = JSON.parse(fs.readFileSync(connectionProfilePath, 'utf8'));
        
        // Створення шлюзу та з'єднання з мережею
        const gateway = new Gateway();
        await gateway.connect(connectionProfile, { 
            wallet, 
            identity: 'admin', 
            discovery: { enabled: true, asLocalhost: true } 
        });
        
        return gateway;
    } catch (error) {
        console.error(`Помилка підключення до мережі: ${error}`);
        return null;
    }
}

// API ендпоінти

// Створення користувача
app.post('/api/v1/users', async (req, res) => {
    try {
        const { id, name, org, roles } = req.body;
        if (!id || !name || !org || !roles) {
            return res.status(400).json({ error: 'Відсутні обов\'язкові параметри' });
        }
        
        // Підключення до мережі
        const gateway = await connectToNetwork();
        if (!gateway) {
            return res.status(500).json({ error: 'Помилка підключення до мережі' });
        }
        
        // Отримання контракту
        const network = await gateway.getNetwork('security-channel');
        const contract = network.getContract('accesscontrol');
        
        // Виклик методу смарт-контракту
        await contract.submitTransaction(
            'CreateUser', 
            id, 
            name, 
            org, 
            JSON.stringify(roles)
        );
        
        // Закриття з'єднання
        gateway.disconnect();
        
        // Відправка відповіді
        return res.status(201).json({ 
            message: 'Користувач успішно створений',
            userId: id
        });
    } catch (error) {
        console.error(`Помилка обробки запиту: ${error}`);
        return res.status(500).json({ error: error.message });
    }
});

// Перевірка доступу
app.post('/api/v1/access/check', async (req, res) => {
    try {
        const { userId, resourceId } = req.body;
        if (!userId || !resourceId) {
            return res.status(400).json({ error: 'Відсутні обов\'язкові параметри' });
        }
        
        // Підключення до мережі
        const gateway = await connectToNetwork();
        if (!gateway) {
            return res.status(500).json({ error: 'Помилка підключення до мережі' });
        }
        
        // Отримання контракту
        const network = await gateway.getNetwork('security-channel');
        const contract = network.getContract('accesscontrol');
        
        // Виклик методу смарт-контракту
        const result = await contract.evaluateTransaction('CheckAccess', userId, resourceId);
        
        // Закриття з'єднання
        gateway.disconnect();
        
        // Аналіз результату
        const accessGranted = JSON.parse(result.toString());
        
        // Відправка відповіді
        return res.status(200).json({ 
            userId, 
            resourceId, 
            accessGranted,
            timestamp: Date.now()
        });
    } catch (error) {
        console.error(`Помилка обробки запиту: ${error}`);
        return res.status(500).json({ error: error.message });
    }
});

// Запис події аудиту
app.post('/api/v1/audit/events', async (req, res) => {
    try {
        const { eventType, actor, resource, action, result, metadata } = req.body;
        if (!eventType || !actor || !resource || !action || !result) {
            return res.status(400).json({ error: 'Відсутні обов\'язкові параметри' });
        }
        
        // Підключення до мережі
        const gateway = await connectToNetwork();
        if (!gateway) {
            return res.status(500).json({ error: 'Помилка підключення до мережі' });
        }
        
        // Отримання контракту
        const network = await gateway.getNetwork('security-channel');
        const contract = network.getContract('securityaudit');
        
        // Виклик методу смарт-контракту
        await contract.submitTransaction(
            'RecordEvent', 
            eventType, 
            actor, 
            resource, 
            action, 
            result,
            JSON.stringify(metadata || {})
        );
        
        // Закриття з'єднання
        gateway.disconnect();
        
        // Відправка відповіді
        return res.status(201).json({ 
            message: 'Подія успішно записана',
            timestamp: Date.now()
        });
    } catch (error) {
        console.error(`Помилка обробки запиту: ${error}`);
        return res.status(500).json({ error: error.message });
    }
});

// Запуск сервера
app.listen(port, () => {
    console.log(`API сервер запущено на порту ${port}`);
});

module.exports = app;