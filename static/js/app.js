let currentUser = null;
let selectedMatch = null;
let selectedTeam = null;
let currentFilter = 'all';

// AUTH 
function showLogin() {
    document.getElementById('login-form').style.display = 'block';
    document.getElementById('register-form').style.display = 'none';
    document.getElementById('login-tab').classList.add('active');
    document.getElementById('register-tab').classList.remove('active');
}

function showRegister() {
    document.getElementById('login-form').style.display = 'none';
    document.getElementById('register-form').style.display = 'block';
    document.getElementById('login-tab').classList.remove('active');
    document.getElementById('register-tab').classList.add('active');
}

async function register() {
    const username = document.getElementById('register-username').value;
    const email = document.getElementById('register-email').value;
    const password = document.getElementById('register-password').value;

    if (!username || !email || !password) {
        alert('Please fill all fields');
        return;
    }

    try {
        const response = await fetch('/api/users/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, email, password })
        });

        const data = await response.json();

        if (response.ok) {
            alert(' Registration successful! You received $100 bonus!');
            currentUser = data.user;
            localStorage.setItem('currentUser', JSON.stringify(currentUser));
            showMainContent();
        } else {
            alert('error ' + data.error);
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

async function login() {
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    if (!email || !password) {
        alert('Please fill all fields');
        return;
    }

    try {
        const response = await fetch('/api/users/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        const data = await response.json();

        if (response.ok) {
            currentUser = data.user;
            localStorage.setItem('currentUser', JSON.stringify(currentUser));
            showMainContent();
        } else {
            alert('error ' + data.error);
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

function logout() {
    currentUser = null;
    localStorage.removeItem('currentUser');
    document.getElementById('auth-section').style.display = 'flex';
    document.getElementById('main-content').style.display = 'none';
    document.getElementById('user-info').style.display = 'none';
}

function showMainContent() {
    document.getElementById('auth-section').style.display = 'none';
    document.getElementById('main-content').style.display = 'block';
    document.getElementById('user-info').style.display = 'flex';
    document.getElementById('username-display').textContent = currentUser.username;
    
    // ПРОВЕРКА НА АДМИНА ПО ПОЛЮ ROLE
    const adminTab = document.getElementById('admin-tab');
    console.log('User role:', currentUser.role); 
    
    if (currentUser.role === 'admin') {
        adminTab.style.display = 'block';
        console.log(' Admin access granted');
    } else {
        adminTab.style.display = 'none';
        console.log(' Regular user');
    }
    
    updateBalance();
    updateUserStats();
    loadMatches();
}

// Вспомогательная функция проверки прав админа
function isAdmin() {
    return currentUser && currentUser.role === 'admin';
}

// BALANCE
async function updateBalance() {
    try {
        const response = await fetch(`/api/users/balance?user_id=${currentUser.id}`);
        const data = await response.json();
        
        if (response.ok) {
            currentUser.balance = data.balance;
            document.getElementById('balance-display').textContent = `$${data.balance.toFixed(2)}`;
            localStorage.setItem('currentUser', JSON.stringify(currentUser));
        }
    } catch (error) {
        console.error('Failed to update balance:', error);
    }
}

async function updateUserStats() {
    try {
        const response = await fetch(`/api/users/profile?user_id=${currentUser.id}`);
        const data = await response.json();
        
        if (response.ok) {
            // Обновляем роль 
            if (data.role) {
                currentUser.role = data.role;
                localStorage.setItem('currentUser', JSON.stringify(currentUser));
            }
            
            // Обновляем win rate в header
            const winRate = data.winRate || 0;
            document.getElementById('winrate-display').textContent = `W/R: ${winRate.toFixed(1)}%`;
            
            // Обновляем статистику
            document.getElementById('total-bets-stat').textContent = data.totalBets || 0;
            document.getElementById('won-bets-stat').textContent = data.wonBets || 0;
            document.getElementById('lost-bets-stat').textContent = data.lostBets || 0;
            document.getElementById('total-wagered-stat').textContent = `$${(data.totalWagered || 0).toFixed(2)}`;
            document.getElementById('total-winnings-stat').textContent = `$${(data.totalWinnings || 0).toFixed(2)}`;
            
            const profitLoss = data.profitLoss || 0;
            const profitLossElem = document.getElementById('profit-loss-stat');
            profitLossElem.textContent = `$${profitLoss.toFixed(2)}`;
            profitLossElem.style.color = profitLoss >= 0 ? '#2ECC71' : '#E74C3C';
        }
    } catch (error) {
        console.error('Failed to update stats:', error);
    }
}

//  NAVIGATION
function showMatches() {
    document.querySelectorAll('.content-section').forEach(s => s.style.display = 'none');
    document.getElementById('matches-section').style.display = 'block';
    setActiveTab('matches-tab');
    loadMatches();
}

function showMyBets() {
    document.querySelectorAll('.content-section').forEach(s => s.style.display = 'none');
    document.getElementById('bets-section').style.display = 'block';
    setActiveTab('bets-tab');
    loadUserBets();
}

function showStats() {
    document.querySelectorAll('.content-section').forEach(s => s.style.display = 'none');
    document.getElementById('stats-section').style.display = 'block';
    setActiveTab('stats-tab');
    updateUserStats();
}

function showAdmin() {
    // Проверка прав админа по полю role
    if (!isAdmin()) {
        alert(' Access denied! Admin panel is only available for administrators.');
        return;
    }
    
    document.querySelectorAll('.content-section').forEach(s => s.style.display = 'none');
    document.getElementById('admin-section').style.display = 'block';
    setActiveTab('admin-tab');
}

function setActiveTab(tabId) {
    document.querySelectorAll('.tabs .tab-btn').forEach(t => t.classList.remove('active'));
    document.getElementById(tabId).classList.add('active');
}

// MATCHES
async function loadMatches() {
    try {
        let url = '/api/matches';
        if (currentFilter !== 'all') {
            url += `?status=${currentFilter}`;
        }

        const response = await fetch(url);
        const matches = await response.json();
        const matchesList = document.getElementById('matches-list');
        
        if (!matches || matches.length === 0) {
            matchesList.innerHTML = '<p style="text-align:center;color:#888;">No matches available</p>';
            return;
        }

        matchesList.innerHTML = matches.map(match => `
            <div class="match-card">
                <div class="match-header">
                    <span class="tournament"> ${match.tournament || 'Tournament'}</span>
                    <span class="match-status status-${match.status}">${match.status.toUpperCase()}</span>
                </div>
                
                <div class="teams">
                    <div class="team">
                        <div class="team-name">${match.teamA.name}</div>
                        <div class="odds">${match.teamA.odds.toFixed(2)}</div>
                        <div style="font-size:0.8em;color:#888;margin-top:5px;">
                            $${match.teamA.totalBets.toFixed(0)} wagered
                        </div>
                    </div>
                    
                    <span class="vs">VS</span>
                    
                    <div class="team">
                        <div class="team-name">${match.teamB.name}</div>
                        <div class="odds">${match.teamB.odds.toFixed(2)}</div>
                        <div style="font-size:0.8em;color:#888;margin-top:5px;">
                            $${match.teamB.totalBets.toFixed(0)} wagered
                        </div>
                    </div>
                </div>
                
                <div style="margin-top:15px;font-size:0.85em;color:#888;">
                    ${match.format} • ${match.betsCount || 0} bets • $${match.totalBetsAmount.toFixed(0)} pool
                </div>
                
                ${match.status === 'upcoming' ? `
                    <button onclick="openBetModal('${match.id}')" class="btn btn-primary" style="width:100%;margin-top:15px;">
                        Place Bet
                    </button>
                ` : ''}
                
                ${match.winner ? `
                    <div style="margin-top:15px;text-align:center;color:#2ECC71;font-weight:bold;">
                        Winner: ${match.winner === 'team_a' ? match.teamA.name : match.teamB.name}
                    </div>
                ` : ''}
            </div>
        `).join('');
    } catch (error) {
        console.error('Failed to load matches:', error);
    }
}

function filterMatches(status) {
    currentFilter = status;
    document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
    event.target.classList.add('active');
    loadMatches();
}

// BETTING 
async function openBetModal(matchId) {
    try {
        const response = await fetch(`/api/matches/details?id=${matchId}`);
        const match = await response.json();

        if (!response.ok) {
            alert('Match not found');
            return;
        }

        selectedMatch = match;
        selectedTeam = null;

        document.getElementById('bet-match-info').innerHTML = `
            <div style="text-align:center;margin-bottom:20px;">
                <h3>${match.teamA.name} vs ${match.teamB.name}</h3>
                <p style="color:#888;">${match.tournament} • ${match.format}</p>
            </div>
        `;

        document.getElementById('bet-team-a-name').textContent = match.teamA.name;
        document.getElementById('bet-team-a-odds').textContent = match.teamA.odds.toFixed(2);
        
        document.getElementById('bet-team-b-name').textContent = match.teamB.name;
        document.getElementById('bet-team-b-odds').textContent = match.teamB.odds.toFixed(2);

        document.getElementById('bet-amount').value = '';
        document.getElementById('potential-win').textContent = '';
        document.getElementById('bet-modal').style.display = 'flex';

        // Click handlers
        document.getElementById('bet-team-a').onclick = () => selectTeam('team_a', match.teamA.odds);
        document.getElementById('bet-team-b').onclick = () => selectTeam('team_b', match.teamB.odds);
    } catch (error) {
        alert('Error loading match: ' + error.message);
    }
}

function selectTeam(team, odds) {
    selectedTeam = { team, odds };
    
    document.getElementById('bet-team-a').classList.remove('selected');
    document.getElementById('bet-team-b').classList.remove('selected');
    
    if (team === 'team_a') {
        document.getElementById('bet-team-a').classList.add('selected');
    } else {
        document.getElementById('bet-team-b').classList.add('selected');
    }
    
    calculatePotentialWin();
}

function calculatePotentialWin() {
    const amount = parseFloat(document.getElementById('bet-amount').value);
    
    if (!selectedTeam || !amount || amount <= 0) {
        document.getElementById('potential-win').textContent = '';
        return;
    }

    const potentialWin = (amount * selectedTeam.odds).toFixed(2);
    const profit = (potentialWin - amount).toFixed(2);
    
    document.getElementById('potential-win').innerHTML = `
        Potential Win: $${potentialWin} (Profit: $${profit})
    `;
}

document.getElementById('bet-amount').addEventListener('input', calculatePotentialWin);

async function placeBet() {
    if (!selectedTeam) {
        alert('Please select a team');
        return;
    }

    const amount = parseFloat(document.getElementById('bet-amount').value);

    if (!amount || amount <= 0) {
        alert('Please enter a valid amount');
        return;
    }

    if (amount > currentUser.balance) {
        alert('Insufficient balance');
        return;
    }

    try {
        const response = await fetch('/api/bets/place', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                user_id: currentUser.id,
                match_id: selectedMatch.id,
                team: selectedTeam.team,
                amount: amount
            })
        });

        const data = await response.json();

        if (response.ok) {
            alert(' Bet placed successfully!');
            closeBetModal();
            updateBalance();
            updateUserStats();
            loadMatches(); // Обновляем матчи 
        } else {
            alert('error' + data.error);
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

function closeBetModal() {
    document.getElementById('bet-modal').style.display = 'none';
    selectedMatch = null;
    selectedTeam = null;
}

//  MY BETS 
async function loadUserBets() {
    try {
        const response = await fetch(`/api/bets/user?user_id=${currentUser.id}`);
        const bets = await response.json();
        const betsList = document.getElementById('bets-list');
        
        console.log('Loaded bets:', bets); 
        
        if (!bets || bets.length === 0) {
            betsList.innerHTML = '<p style="text-align:center;color:#888;">No bets yet</p>';
            return;
        }

        // ИСПРАВЛЕНО: объявляем переменные ДО использования в map
        betsList.innerHTML = bets.map(bet => {
            // Определяем класс статуса
            let statusClass = 'pending';
            if (bet.status === 'won') statusClass = 'won';
            if (bet.status === 'lost') statusClass = 'lost';
            if (bet.status === 'cancelled') statusClass = 'pending';
            
            // Определяем эмодзи статуса
            let statusEmoji = '⏳';
            if (bet.status === 'won') statusEmoji = '✅';
            if (bet.status === 'lost') statusEmoji = '❌';
            if (bet.status === 'cancelled') statusEmoji = '🚫';
            
            return `
                <div class="bet-card ${statusClass}">
                    <div style="display:flex;justify-content:space-between;margin-bottom:10px;">
                        <strong>${bet.matchInfo || 'Match'}</strong>
                        <span>${statusEmoji} ${bet.status.toUpperCase()}</span>
                    </div>
                    <div style="font-size:0.9em;color:#888;">
                        Team: ${bet.team === 'team_a' ? 'A' : 'B'} •
                        Amount: $${bet.amount.toFixed(2)} •
                        Odds: ${bet.odds.toFixed(2)}
                    </div>
                    <div style="margin-top:10px;font-weight:bold;">
                        ${bet.status === 'won' ?
                            ` Won: $${bet.actualWinnings.toFixed(2)}` :
                            bet.status === 'lost' ?
                            `Lost: $${bet.amount.toFixed(2)}` :
                            bet.status === 'cancelled' ?
                            `Refunded: $${bet.amount.toFixed(2)}` :
                            `Potential Win: $${bet.potentialWin.toFixed(2)}`
                        }
                    </div>
                    <div style="font-size:0.8em;color:#666;margin-top:5px;">
                        ${new Date(bet.createdAt).toLocaleString()}
                    </div>
                </div>
            `;
        }).join('');
        
        console.log(' Bets loaded successfully');
    } catch (error) {
        console.error('Failed to load bets:', error);
        document.getElementById('bets-list').innerHTML = 
            '<p style="text-align:center;color:#e74c3c;"> Error loading bets. Check console for details.</p>';
    }
}

// ADMIN 
async function createMatch() {
    // Проверка прав админа
    if (!isAdmin()) {
        alert(' Access denied! Admin privileges required.');
        return;
    }
    
    const teamAId = document.getElementById('team-a-id').value;
    const teamBId = document.getElementById('team-b-id').value;
    const tournament = document.getElementById('tournament').value;
    const format = document.getElementById('format').value;
    const startTime = document.getElementById('start-time').value;

    if (!teamAId || !teamBId || !tournament || !startTime) {
        alert('Please fill all fields');
        return;
    }

    try {
        const response = await fetch('/api/matches', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                teamAId: teamAId,
                teamBId: teamBId,
                tournament: tournament,
                format: format,
                startTime: new Date(startTime).toISOString()
            })
        });

        const data = await response.json();

        if (response.ok) {
            alert(' Match created successfully!');
            document.getElementById('team-a-id').value = '';
            document.getElementById('team-b-id').value = '';
            document.getElementById('tournament').value = '';
            document.getElementById('start-time').value = '';
            loadMatches();
        } else {
            alert(' ' + data.error);
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

async function settleMatch() {
    // Проверка прав админа
    if (!isAdmin()) {
        alert(' Access denied! Admin privileges required.');
        return;
    }
    
    const matchId = document.getElementById('settle-match-id').value;
    const winner = document.getElementById('winner').value;

    if (!matchId) {
        alert('Please enter match ID');
        return;
    }

    try {
        const response = await fetch('/api/bets/settle', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                match_id: matchId,
                winner: winner
            })
        });

        const data = await response.json();

        if (response.ok) {
            alert(' Match settled successfully!');
            document.getElementById('settle-match-id').value = '';
            loadMatches();
            updateBalance();
            updateUserStats();
        } else {
            alert('error ' + data.error);
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

async function addBalance() {
    // Проверка прав админа
    if (!isAdmin()) {
        alert(' Access denied! Admin privileges required.');
        return;
    }
    
    const amount = parseFloat(document.getElementById('add-balance-amount').value);

    if (!amount || amount <= 0) {
        alert('Please enter a valid amount');
        return;
    }

    try {
        const response = await fetch('/api/users/add-balance', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                user_id: currentUser.id,
                amount: amount
            })
        });

        const data = await response.json();

        if (response.ok) {
            alert(' Balance added successfully!');
            document.getElementById('add-balance-amount').value = '';
            updateBalance();
        } else {
            alert('error ' + data.error);
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

// AUTO-REFRESH 
setInterval(() => {
    if (currentUser && document.getElementById('matches-section').style.display === 'block') {
        loadMatches();
    }
}, 30000); // Обновляем матчи каждые 30 секунд

//INIT
window.onload = () => {
    const storedUser = localStorage.getItem('currentUser');
    if (storedUser) {
        currentUser = JSON.parse(storedUser);
        showMainContent();
    }
};