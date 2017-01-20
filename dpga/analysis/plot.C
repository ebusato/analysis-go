
Double_t fitf(Double_t *v, Double_t *par)
{
	Double_t fitval = 0; 
	if(v[0] >= par[0]) {
		fitval = par[1] + par[2]*-1*TMath::Exp(-(v[0] - par[3])/par[4]) + par[5]*TMath::Exp(-(v[0] - par[3])/par[6]);
	}
	return fitval;
}

void plot(TString name, int evt)
{
	gStyle->SetPadGridX(true);
	gStyle->SetPadGridY(true);
	
	TFile* f = new TFile(name, "read");
	TTree* tree = (TTree*) f->Get("tree");
	
	TF1 *fExp = new TF1("fExp","TMath::Exp([0] - [1]*x)");
	
	// Make plot
	TCanvas* c1 = new TCanvas("c1","c1",900,900);
	c1->Divide(2,2);
	c1->cd(1);
	tree->Draw("T30[0] - T30[1]>>hCRT(200,-30,30)");
	c1->cd(2);
	tree->Draw("Pulse[1] : SampleTimes", Form("Evt == %i", evt),"goff");
	TGraph *g1 = new TGraph(999,tree->GetV2(),tree->GetV1());
	g1->SetMarkerColor(kRed);
	g1->SetMarkerStyle(8);
	g1->SetMarkerSize(0.5);
	g1->Fit("fExp", "", "", 75, 200);
	TF1* fg1 = g1->GetFunction("fExp");
	fg1->SetLineColor(kOrange+1);
	fg1->SetLineWidth(4);
	Double_t par1 = fExp->GetParameter(1);
	Double_t decayTime0 = 0;
	if(par1 != 0) {
		decayTime0 = 1/par1;
	}
	cout << "decayTime0 = " << decayTime0 << " ns" << endl;
	tree->Draw("Pulse[0]*Ampl[1]/Ampl[0] : SampleTimes", Form("Evt == %i", evt),"same");
	TGraph *g2 = new TGraph(999,tree->GetV2(),tree->GetV1());
	g2->SetMarkerColor(kBlue);
	g2->SetMarkerStyle(8);
	g2->SetMarkerSize(0.5);
	g2->Fit("fExp", "", "", 75, 200);
	TF1* fg2 = g2->GetFunction("fExp");
	fg2->SetLineColor(kGreen+1);
	fg2->SetLineWidth(4);
	par1 = fExp->GetParameter(1);
	Double_t decayTime1 = 0;
	if(par1 != 0) {
		decayTime1 = 1/par1;
	}
	cout << "decayTime1 = " << decayTime1 << " ns" << endl;
	TMultiGraph* mg = new TMultiGraph();
	mg->Add(g1);
	mg->Add(g2);
	mg->Draw("ap");
	
	// Retrieve values and print them on plot
	tree->Draw("Ampl[1]", Form("Evt == %i", evt),"goff");
	amplitude1 = *tree->GetV1();
	tree->Draw("Ampl[0]", Form("Evt == %i", evt),"goff");
	amplitude0 = *tree->GetV1();
	TLine* l10 = new TLine(0, amplitude1*0.1, 200, amplitude1*0.1);
	l10->SetLineWidth(2);
	l10->SetLineColor(kRed);
	l10->Draw("same");
	TLine* l20 = new TLine(0, amplitude1*0.2, 200, amplitude1*0.2);
	l20->SetLineWidth(2);
	l20->Draw("same");
	TLine* l30 = new TLine(0, amplitude1*0.3, 200, amplitude1*0.3);
	l30->SetLineWidth(2);
	l30->Draw("same");
	TLine* l80 = new TLine(0, amplitude1*0.8, 200, amplitude1*0.8);
	l80->SetLineWidth(2);
	l80->Draw("same");
	TLine* l90 = new TLine(0, amplitude1*0.9, 200, amplitude1*0.9);
	l90->SetLineWidth(2);
	l90->SetLineColor(kRed);
	l90->Draw("same");
	tree->Draw("T30[0] - T30[1]", Form("Evt == %i", evt),"goff");
	TLatex* la1 = new TLatex(110, amplitude1 * 0.6, Form("#Delta t = %f ns", *tree->GetV1()));
	la1->Draw();
	TLatex* la2 = new TLatex(110, amplitude1 * 0.8, Form("ampl[0]=%f", amplitude0));
	la2->SetTextColor(kBlue);
	la2->Draw();
	TLatex* la3 = new TLatex(110, amplitude1 * 0.7, Form("ampl[1]=%f", amplitude1));
	la3->SetTextColor(kRed);
	la3->Draw();
	TLatex* la4 = new TLatex(110, amplitude1 * 0.5, Form("Decay time 0=%f", decayTime0));
	la4->SetTextColor(kRed);
	la4->Draw();
	TLatex* la5 = new TLatex(110, amplitude1 * 0.4, Form("Decay time 1=%f", decayTime1));
	la5->SetTextColor(kRed);
	la5->Draw();
	
}