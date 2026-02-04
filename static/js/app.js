let currentUser = null;
let currentMatchId = null;
let selectedTeam = null;
let currentFilter = 'all';

function showLogin() {
    document.getElementById('login-form').style.display = 'flex';
    document.getElementById('register-form').style.display = 'none';
    document.getElementById('login-tab').classList.add('active');
    document.getElementById('register-tab').classList.remove('active');
}

function showRegister() {
    document.getElementById('login-form').style.display = 'none';
    document.getElementById('register-form').style.display = 'flex';
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
            body: JSON.stringify({ username, email, password, balance: 1000 })
        });

        const data = await response.json();
        
        if (response.ok) {
            alert('Registration successful! Please login.');
            showLogin();
        } else {
            alert(data.error || 'Registration failed');
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
            alert(data.error || 'Login failed');
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

function logout() {
    currentUser = null;
    localStorage.removeItem('currentUser');
    document.getElementById('auth-section').style.display = 'block';
    document.getElementById('main-content').style.display = 'none';
    document.getElementById('user-info').style.display = 'none';
}

function showMainContent() {
    document.getElementById('auth-section').style.display = 'none';
    document.getElementById('main-content').style.display = 'block';
    document.getElementById('user-info').style.display = 'flex';
    document.getElementById('username-display').textContent = currentUser.username;
    updateBalance();
    loadMatches();

    setInterval(() => {
        if (document.getElementById('matches-section').style.display !== 'none') {
            loadMatches();
        }
    }, 10000);
}

async function updateBalance() {
    try {
        const response = await fetch(`/api/users/balance?id=${currentUser.id}`);
        const data = await response.json();
        
        if (response.ok) {
            currentUser.balance = data.balance;
            document.getElementById('balance-display').textContent = `Balance: $${data.balance.toFixed(2)}`;
        }
    } catch (error) {
        console.error('Error updating balance:', error);
    }
}

function showMatches() {
    document.getElementById('matches-section').style.display = 'block';
    document.getElementById('bets-section').style.display = 'none';
    document.getElementById('admin-section').style.display = 'none';
    
    document.getElementById('matches-tab').classList.add('active');
    document.getElementById('bets-tab').classList.remove('active');
    document.getElementById('admin-tab').classList.remove('active');
    
    loadMatches();
}

function showMyBets() {
    document.getElementById('matches-section').style.display = 'none';
    document.getElementById('bets-section').style.display = 'block';
    document.getElementById('admin-section').style.display = 'none';
    
    document.getElementById('matches-tab').classList.remove('active');
    document.getElementById('bets-tab').classList.add('active');
    document.getElementById('admin-tab').classList.remove('active');
    
    loadMyBets();
}

function showAdmin() {
    document.getElementById('matches-section').style.display = 'none';
    document.getElementById('bets-section').style.display = 'none';
    document.getElementById('admin-section').style.display = 'block';
    
    document.getElementById('matches-tab').classList.remove('active');
    document.getElementById('bets-tab').classList.remove('active');
    document.getElementById('admin-tab').classList.add('active');
}

async function loadMatches() {
    try {
        const url = currentFilter === 'all' 
            ? '/api/matches' 
            : `/api/matches?status=${currentFilter}`;
            
        const response = await fetch(url);
        const matches = await response.json();
        
        displayMatches(matches);
    } catch (error) {
        console.error('Error loading matches:', error);
    }
}

function filterMatches(filter) {
    currentFilter = filter;
    loadMatches();
}

function displayMatches(matches) {
    const matchesList = document.getElementById('matches-list');
    
    if (!matches || matches.length === 0) {
        matchesList.innerHTML = '<p>No matches available</p>';
        return;
    }
    
    matchesList.innerHTML = matches.map(match => `
        <div class="match-card">
            <div class="match-header">
                <span class="tournament">${match.tournament}</span>
                <span class="status ${match.status}">${match.status.toUpperCase()}</span>
            </div>
            
            <div class="teams">
                <div class="team">
                    <div class="team-name">${match.team_a.name}</div>
                    <div class="odds">${match.team_a.odds.toFixed(2)}</div>
                    <div style="font-size: 12px; color: #666;">Total: $${match.team_a.total_bets.toFixed(2)}</div>
                </div>
                
                <div class="vs">VS</div>
                
                <div class="team">
                    <div class="team-name">${match.team_b.name}</div>
                    <div class="odds">${match.team_b.odds.toFixed(2)}</div>
                    <div style="font-size: 12px; color: #666;">Total: $${match.team_b.total_bets.toFixed(2)}</div>
                </div>
            </div>
            
            <div class="match-footer">
                <span class="match-time">${new Date(match.start_time).toLocaleString()}</span>
                ${match.status === 'upcoming' 
                    ? `<button onclick="openBetModal('${match.id}')" class="btn btn-primary">Place Bet</button>`
                    : match.winner 
                        ? `<span style="color: #065f46; font-weight: bold;">Winner: ${match.winner === 'team_a' ? match.team_a.name : match.team_b.name}</span>`
                        : ''
                }
            </div>
        </div>
    `).join('');
}

async function openBetModal(matchId) {
    currentMatchId = matchId;
    selectedTeam = null;
    
    try {
        const response = await fetch(`/api/matches/details?id=${matchId}`);
        const match = await response.json();
        
        document.getElementById('bet-team-a-name').textContent = match.team_a.name;
        document.getElementById('bet-team-a-odds').textContent = match.team_a.odds.toFixed(2);
        document.getElementById('bet-team-b-name').textContent = match.team_b.name;
        document.getElementById('bet-team-b-odds').textContent = match.team_b.odds.toFixed(2);
        
        document.getElementById('bet-team-a').onclick = () => selectTeam('team_a', match.team_a.odds);
        document.getElementById('bet-team-b').onclick = () => selectTeam('team_b', match.team_b.odds);
        
        document.getElementById('bet-amount').value = '';
        document.getElementById('potential-win').textContent = '';
        
        document.getElementById('bet-modal').style.display = 'block';
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
    
    if (selectedTeam && amount > 0) {
        const potentialWin = amount * selectedTeam.odds;
        document.getElementById('potential-win').textContent = 
            `Potential Win: $${potentialWin.toFixed(2)}`;
    } else {
        document.getElementById('potential-win').textContent = '';
    }
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
                match_id: currentMatchId,
                selected_team: selectedTeam.team,
                amount: amount
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            alert('Bet placed successfully!');
            closeBetModal();
            updateBalance();
            loadMatches();
        } else {
            alert(data.error || 'Failed to place bet');
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

function closeBetModal() {
    document.getElementById('bet-modal').style.display = 'none';
}

async function loadMyBets() {
    try {
        const response = await fetch(`/api/bets/user?user_id=${currentUser.id}`);
        const bets = await response.json();
        
        displayBets(bets);
    } catch (error) {
        console.error('Error loading bets:', error);
    }
}

function displayBets(bets) {
    const betsList = document.getElementById('bets-list');
    
    if (!bets || bets.length === 0) {
        betsList.innerHTML = '<p>No bets placed yet</p>';
        return;
    }
    
    betsList.innerHTML = bets.map(bet => `
        <div class="bet-card">
            <div class="bet-header">
                <div>
                    <strong>${bet.match_info.team_a.name} vs ${bet.match_info.team_b.name}</strong>
                    <div style="font-size: 12px; color: #666;">${bet.match_info.tournament}</div>
                </div>
                <span class="bet-status ${bet.status}">${bet.status.toUpperCase()}</span>
            </div>
            
            <div class="bet-info">
                <div class="bet-info-item">
                    <span class="bet-info-label">Selected:</span>
                    <span class="bet-info-value">${bet.selected_team === 'team_a' ? bet.match_info.team_a.name : bet.match_info.team_b.name}</span>
                </div>
                <div class="bet-info-item">
                    <span class="bet-info-label">Amount:</span>
                    <span class="bet-info-value">$${bet.amount.toFixed(2)}</span>
                </div>
                <div class="bet-info-item">
                    <span class="bet-info-label">Odds:</span>
                    <span class="bet-info-value">${bet.odds.toFixed(2)}</span>
                </div>
                <div class="bet-info-item">
                    <span class="bet-info-label">Potential Win:</span>
                    <span class="bet-info-value">$${bet.potential_win.toFixed(2)}</span>
                </div>
                ${bet.win_amount > 0 ? `
                <div class="bet-info-item">
                    <span class="bet-info-label">Won:</span>
                    <span class="bet-info-value" style="color: #065f46;">$${bet.win_amount.toFixed(2)}</span>
                </div>
                ` : ''}
            </div>
            
            <div style="margin-top: 10px; font-size: 12px; color: #666;">
                Placed: ${new Date(bet.created_at).toLocaleString()}
            </div>
        </div>
    `).join('');
}

async function createMatch() {
    const teamAName = document.getElementById('team-a-name').value;
    const teamALogo = document.getElementById('team-a-logo').value;
    const teamBName = document.getElementById('team-b-name').value;
    const teamBLogo = document.getElementById('team-b-logo').value;
    const tournament = document.getElementById('tournament').value;
    const format = document.getElementById('format').value;
    const startTime = document.getElementById('start-time').value;
    
    if (!teamAName || !teamBName || !tournament || !startTime) {
        alert('Please fill all required fields');
        return;
    }
    
    try {
        const response = await fetch('/api/matches', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                team_a_name: teamAName,
                team_a_logo: teamALogo,
                team_b_name: teamBName,
                team_b_logo: teamBLogo,
                tournament: tournament,
                format: format,
                start_time: new Date(startTime).toISOString()
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            alert('Match created successfully!');

            document.getElementById('team-a-name').value = '';
            document.getElementById('team-a-logo').value = '';
            document.getElementById('team-b-name').value = '';
            document.getElementById('team-b-logo').value = '';
            document.getElementById('tournament').value = '';
            document.getElementById('start-time').value = '';
            loadMatches();
        } else {
            alert(data.error || 'Failed to create match');
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

async function settleMatch() {
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
            alert('Match settled successfully!');
            document.getElementById('settle-match-id').value = '';
            loadMatches();
        } else {
            alert(data.error || 'Failed to settle match');
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

async function addBalance() {
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
            alert('Balance added successfully!');
            document.getElementById('add-balance-amount').value = '';
            updateBalance();
        } else {
            alert(data.error || 'Failed to add balance');
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

window.onload = () => {
    const storedUser = localStorage.getItem('currentUser');
    if (storedUser) {
        currentUser = JSON.parse(storedUser);
        showMainContent();
    }
};
